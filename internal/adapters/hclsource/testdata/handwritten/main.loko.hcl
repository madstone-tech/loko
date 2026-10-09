# Hand-written fixture for feature 015. Deliberately NOT canonical: misaligned
# equals signs, comment styles of every kind, irregular blank lines, and an
# order that is neither alphabetical nor grouped by kind. Editing tools must
# leave every byte outside the edited declaration exactly as it is.

project "handwritten" {
  description = "Byte-preservation fixture"
}



// The storefront. Owned by commerce.
system "shop" {
  description="Storefront"   # trailing comment
  owner   = "commerce"
}

/* A block comment
   spanning lines */
container "web" {
  system = system.shop
  technology    = "React"
  tags = ["edge",   "public"]

  uses "api" {
    target      = container.api
    description = "Calls the API"
  }
}

person "customer" {
  # nested comment
  uses "browse" { target = container.web }
}

container "api" {
  system      = system.shop
  description = "Storefront API"
  // comment between attributes
  technology = "Go"

  uses "charge" {
    target = container.gateway
  }
}

system "payments" {
  description = "Payments"
}

# Separated from the next block by a blank line: belongs to no declaration.

container "gateway" {
  system = system.payments
  uses   "acquire" { target = external.bank }
}
// attached comment for bank
external "bank" {
  description = "Acquiring bank"
}

component "handler" {
  container = container.api
}
