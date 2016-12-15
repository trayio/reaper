# nginx-production group
"nginx-production" {
  count = 3
  age = 72 # 3 days
  region = "east"
}

"nginx-staging" {
  count = 1
  age = 5
  region = "west"
}
