project "edge-cases" {}

# The most file-system-hostile label the grammar accepts (arch.ValidName):
# a leading underscore, mixed case and a hyphen. "_" must be escaped.
system "_Odd_name-1" {
  uses "self" {
    target      = system._Odd_name-1
    description = "Calls itself"
  }
}

system "a" {
  uses "b" { target = system.b }
}
system "b" {
  uses "c" { target = system.c }
}
system "c" {
  uses "a" { target = system.a }
}

system "hollow" {
  description = "No containers"
}
