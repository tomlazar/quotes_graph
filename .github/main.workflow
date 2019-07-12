workflow "Build and Deploy on Release" {
  on = "release"
  resolves = ["Notify Slack", "Deploy Changes"]
}

action "Notify Slack" {
  uses = "Ilshidur/action-slack@4f95940e640ecbb8d2bc330aaeea05256814467a"
  secrets = ["SLACK_WEBHOOK"]
  args = "New version of Slack bot Deployed"
  
}

action "Deploy Changes" {
  uses = "maddox/actions/ssh@master"
  secrets = ["PRIVATE_KEY", "PUBLIC_KEY", "USER", "HOST", "PORT"]
  args = "cd $HOME/deploy/quotes_graph && git clone && docker-compose up --build"
  needs = ["Notify Slack"]
}
