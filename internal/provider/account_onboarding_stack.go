package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type accountOnboardingResult struct {
	accountId        string
	onboardingStatus string
	errorMessage     error
	stackId          string
}

type accountOnboardingStackInput struct {
	accountId           string
	endpointMode        string
	decryptionCert      string
	cloudwatchNamespace string
	cloudwatchLogGroup  string
	auditLogGroup       string
	kinesisFirehose     string
	s3Bucket            string
	snsTopicArn         string
	trustedAccount      string
	externalId          string
	cftRoleName         string
	onboardingCft       string
	stackId             string
	region              string
	profile             string
}

// CloudFormationClient returns a AWS cloudformation client by assuming the CFT role in the specified account.
func CloudFormationClient(ctx context.Context, accountId, cftRoleName, region, profile string) (*cloudformation.Client, error) {
	cftRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, cftRoleName)
	options := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if profile != "" {
		options = append(options, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return nil, err
	}
	stsClient := sts.NewFromConfig(cfg)
	assumeRoleOutput, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         PtrToString(cftRoleArn),
		RoleSessionName: PtrToString("test"),
	})
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return nil, err
	}
	creds := Creds{
		Credentials: assumeRoleOutput.Credentials,
	}
	svc := cloudformation.NewFromConfig(aws.Config{Credentials: creds, Region: region})
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return nil, err
	}
	return svc, nil
}

func FindStackByName(ctx context.Context, name string, nextToken *string,
	cfrClient *cloudformation.Client) (string, error) {
	listStacksInput := &cloudformation.ListStacksInput{
		NextToken: nextToken,
	}
	response, err := cfrClient.ListStacks(ctx, listStacksInput)
	if err != nil {
		return "", err
	}
	stacks := response.StackSummaries
	if len(stacks) == 0 {
		return "", nil
	}
	for _, stack := range stacks {
		if *stack.StackName == name {
			return *stack.StackId, nil
		}
	}
	if response.NextToken == nil {
		return "", nil
	}
	return FindStackByName(ctx, name, response.NextToken, cfrClient)
}

func CreateAccountOnboardingStack(ctx context.Context, input accountOnboardingStackInput) (string, error) {
	cfrClient, err := CloudFormationClient(ctx, input.accountId, input.cftRoleName, input.region, input.profile)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return "", err
	}
	tflog.Info(ctx, "creating stack")
	createStackInput := &cloudformation.CreateStackInput{
		StackName:    PtrToString(DefaultCFTStackName),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
		Parameters: []types.Parameter{
			{
				ParameterKey:   PtrToString("TrustedAccount"),
				ParameterValue: &input.trustedAccount,
			},
			{
				ParameterKey:   PtrToString("ExternalId"),
				ParameterValue: &input.externalId,
			},
			{
				ParameterKey:   PtrToString("SNSTopicArn"),
				ParameterValue: &input.snsTopicArn,
			},
			{
				ParameterKey:   PtrToString("EndpointMode"),
				ParameterValue: &input.endpointMode,
			},
			{
				ParameterKey:   PtrToString("DecryptionCertificate"),
				ParameterValue: &input.decryptionCert,
			},
			{
				ParameterKey:   PtrToString("CloudwatchNamespace"),
				ParameterValue: &input.cloudwatchNamespace,
			},
			{
				ParameterKey:   PtrToString("CloudwatchLog"),
				ParameterValue: &input.cloudwatchLogGroup,
			},
			{
				ParameterKey:   PtrToString("AuditLogGroup"),
				ParameterValue: &input.auditLogGroup,
			},
			{
				ParameterKey:   PtrToString("KinesisFirehose"),
				ParameterValue: &input.kinesisFirehose,
			},
			{
				ParameterKey:   PtrToString("S3Bucket"),
				ParameterValue: &input.s3Bucket,
			},
		},
		TemplateBody: PtrToString(input.onboardingCft),
		OnFailure:    types.OnFailureDelete,
	}
	createStackResponse, err := cfrClient.CreateStack(ctx, createStackInput)
	tflog.Info(ctx, "creating stack response: %v", createStackResponse)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return "", err
	}
	stackId := *createStackResponse.StackId
	err = WaitForStackDeployment(ctx, stackId, cfrClient, input.accountId)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return "", err
	}
	return stackId, nil
}

func WaitForStackDeployment(ctx context.Context, stackId string, svc *cloudformation.Client, accountId string) error {
	failedStatus := []types.StackStatus{
		types.StackStatusCreateFailed,
		types.StackStatusRollbackFailed,
		types.StackStatusRollbackComplete,
		types.StackStatusRollbackInProgress,
		types.StackStatusDeleteFailed,
		types.StackStatusDeleteComplete,
	}
	for i := 0; i <= 10; i++ {
		input := &cloudformation.DescribeStacksInput{
			StackName: PtrToString(stackId),
		}
		res, err := svc.DescribeStacks(ctx, input)
		if err != nil {
			return err
		}
		if len(res.Stacks) == 0 {
			return fmt.Errorf("did not get any stacks for id: %s", stackId)
		}
		stackStatus := res.Stacks[0].StackStatus
		for _, status := range failedStatus {
			if stackStatus == status {
				return fmt.Errorf("failed to create cft stack, status: %s, account ID: %s", stackStatus, accountId)
			}
		}
		if stackStatus == types.StackStatusCreateComplete {
			return nil
		}
		time.Sleep(30 * time.Second)
	}
	return nil
}

func WaitForStackDeletion(ctx context.Context, svc *cloudformation.Client, stackId string, accountId string) error {
	failedStatus := []types.StackStatus{
		types.StackStatusDeleteFailed,
		types.StackStatusRollbackFailed,
		types.StackStatusRollbackComplete,
		types.StackStatusRollbackInProgress,
	}
	for i := 0; i <= 10; i++ {
		input := &cloudformation.DescribeStacksInput{
			StackName: PtrToString(stackId),
		}
		res, err := svc.DescribeStacks(ctx, input)
		if err != nil {
			return err
		}
		tflog.Info(ctx, "stacks: %v", res)
		if len(res.Stacks) == 0 {
			return nil
		}
		stackStatus := res.Stacks[0].StackStatus
		for _, status := range failedStatus {
			if stackStatus == status {
				return fmt.Errorf("failed to delete cft stack, status: %s, accountId: %s", stackStatus, accountId)
			}
		}
		if stackStatus == types.StackStatusDeleteComplete {
			return nil
		}
		time.Sleep(30 * time.Second)
	}
	return nil
}

func DeleteStack(ctx context.Context, input accountOnboardingStackInput) error {
	cfrClient, err := CloudFormationClient(ctx, input.accountId, input.cftRoleName, input.region, input.profile)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return err
	}
	deleteStackInput := &cloudformation.DeleteStackInput{
		StackName: PtrToString(input.stackId),
	}
	_, err = cfrClient.DeleteStack(ctx, deleteStackInput)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return err
	}
	err = WaitForStackDeletion(ctx, cfrClient, input.stackId, input.accountId)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return err
	}
	return nil
}

func ReadStack(ctx context.Context, input accountOnboardingStackInput) (string, error) {
	cfrClient, err := CloudFormationClient(ctx, input.accountId, input.cftRoleName, input.region, input.profile)
	if err != nil {
		tflog.Info(ctx, "error: %s", err)
		return "", err
	}
	describeStacksInput := &cloudformation.DescribeStacksInput{
		StackName: PtrToString(input.stackId),
	}
	res, err := cfrClient.DescribeStacks(ctx, describeStacksInput)
	if err != nil {
		return "", err
	}
	if len(res.Stacks) == 0 {
		return "", fmt.Errorf("failed to find any stack with ID: %s", input.stackId)
	}
	stackStatus := res.Stacks[0].StackStatus
	return string(stackStatus), nil
}

// Resource.
func resourceAccountOnboardingStack() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for Account Onboarding.",

		CreateContext: createAccountOnboardingStack,
		DeleteContext: deleteAccountOnboardingStack,
		ReadContext:   readAccountOnboardingStack,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: accountOnboardingStackSchema(true, nil),
	}
}

func createAccountOnboardingStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	mpRegion := svc.GetMPRegion(ctx)
	profile := svc.GetProfile(ctx)
	accountId := d.Get("account_id").(string)
	stackInput := accountOnboardingStackInput{
		auditLogGroup:       d.Get("auditlog_group").(string),
		cftRoleName:         d.Get("cft_role_name").(string),
		endpointMode:        d.Get("endpoint_mode").(string),
		decryptionCert:      d.Get("decryption_cert").(string),
		cloudwatchLogGroup:  d.Get("cloudwatch_log_group").(string),
		kinesisFirehose:     d.Get("kinesis_firehose").(string),
		cloudwatchNamespace: d.Get("cloudwatch_namespace").(string),
		s3Bucket:            d.Get("s3_bucket").(string),
		onboardingCft:       d.Get("onboarding_cft").(string),
		trustedAccount:      d.Get("trusted_account").(string),
		externalId:          d.Get("external_id").(string),
		snsTopicArn:         d.Get("sns_topic_arn").(string),
		accountId:           accountId,
		region:              mpRegion,
		profile:             profile,
	}
	stackId, err := CreateAccountOnboardingStack(ctx, stackInput)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(stackId)
	d.Set("account_id", accountId)
	d.Set("stack_id", stackId)
	return nil
}

func deleteAccountOnboardingStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	mpRegion := svc.GetMPRegion(ctx)
	accountId := d.Get("account_id").(string)
	stackInput := accountOnboardingStackInput{
		cftRoleName: d.Get("cft_role_name").(string),
		stackId:     d.Get("stack_id").(string),
		accountId:   accountId,
		region:      mpRegion,
	}
	err := DeleteStack(ctx, stackInput)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func readAccountOnboardingStack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	mpRegion := svc.GetMPRegion(ctx)
	accountId := d.Get("account_id").(string)
	stackId := d.Get("stack_id").(string)
	stackInput := accountOnboardingStackInput{
		cftRoleName: d.Get("cft_role_name").(string),
		stackId:     stackId,
		accountId:   accountId,
		region:      mpRegion,
	}
	stackStatus, err := ReadStack(ctx, stackInput)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(stackId)
	d.Set("stack_status", stackStatus)
	d.Set("account_id", accountId)
	return nil
}

// Data source for a single Account onboarding stack.
func dataSourceAccountOnboardingStack() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving accont onboarding stack information.",

		ReadContext: readAccountOnboardingStackDataSource,

		Schema: accountOnboardingStackSchema(false, nil),
	}
}

func readAccountOnboardingStackDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	mpRegion := svc.GetMPRegion(ctx)
	accountId := d.Get("account_id").(string)
	stackId := d.Get("stack_id").(string)
	stackInput := accountOnboardingStackInput{
		cftRoleName: d.Get("cft_role_name").(string),
		stackId:     stackId,
		accountId:   accountId,
		region:      mpRegion,
	}
	stackStatus, err := ReadStack(ctx, stackInput)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(stackId)
	d.Set("stack_status", stackStatus)
	d.Set("account_id", accountId)
	return nil
}

// Schema handling.
func accountOnboardingStackSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		"account_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The account IDs",
		},
		"stack_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "ID of the account onboarding CFT stack",
		},
		"endpoint_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Controls whether cloud NGFW will create firewall endpoints automatitically in customer subnets",
		},
		"decryption_cert": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Description: `The CloudNGFW can decrypt inbound and outbound traffic by providing a
						  certificate stored in secret Manager.
			 			  The role allows the service to access a certificate configured in the rulestack.
			 			  Only certificated tagged with PaloAltoCloudNGFW can be accessed`,
		},
		"cloudwatch_namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Cloudwatch Namespace",
		},
		"cloudwatch_log_group": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Cloudwatch Log Group",
		},
		"auditlog_group": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Audit Log Group Name",
		},
		"kinesis_firehose": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Kinesis Firehose for logging",
		},
		"s3_bucket": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "S3 Bucket Name for Logging. Logging roles provide access to create log contents in this bucket.",
		},
		"cft_role_name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Role name to run the account onboarding CFT in each account to be onboarded.",
		},
		"onboarding_cft": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Role name to run the account onboarding CFT in each account to be onboarded.",
		},
		"stack_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "Status of the account onboarding CFT stack.",
		},
		"trusted_account": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "PANW Cloud NGFW trusted account Id",
		},
		"external_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "External Id of the onboarded account",
		},
		"sns_topic_arn": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			Description: "SNS topic ARN to publish the role ARNs",
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}
	return ans
}
