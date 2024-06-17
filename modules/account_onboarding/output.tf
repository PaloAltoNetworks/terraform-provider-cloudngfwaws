output "onboarded_accounts" {
  value = zipmap([ for account in cloudngfwaws_account_onboarding.account_onboarding: account.account_id ], [ for account in cloudngfwaws_account_onboarding.account_onboarding: account.onboarding_status ])
}
