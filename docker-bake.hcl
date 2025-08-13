
# docker-bake.hcl for building multiple images
# Example for a multi-service project

group "default" {
  targets = ["server"]
}

target "server" {
  context = "."
  dockerfile = "Dockerfile"
  tags = ["your-registry/ultra-scalable-monolith:latest"]
  args = {
    GO_VERSION = "1.24.3"
  }
}
