project "case-collision" {}

system "s" {}

# Container views are named after their container, so these two produce
# diagrams/container-Api.* and diagrams/container-api.*, which collide on a
# case-insensitive file system (FR-028).
container "Api" { system = system.s }
container "api" { system = system.s }

component "a1" { container = container.Api }
component "a2" { container = container.api }
