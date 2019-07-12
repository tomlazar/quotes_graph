workflow "Build and Deploy on Release" {
  on = "release"
  resolves = ["maddox/actions/ssh@master"]
}

action "maddox/actions/ssh@master" {
  uses = "maddox/actions/ssh@master"
  secrets = ["PRIVATE_KEY", "PUBLIC_KEY", "USER", "HOST", "PORT"]
  args = "cd $HOME/deploy/quotes_graph && git clone && docker-compose up --build"
}
