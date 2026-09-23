project "p" {}

system "a" {
  colour = "red"
}

system "b" {
  description = trimspace("  x  ")
}

for_each "nope" {}

variable "also-nope" {}
