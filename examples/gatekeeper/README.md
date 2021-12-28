# Gatekeeper examples

This directory contains example YAML to configure [Gatekeeper](https://github.com/open-policy-agent/gatekeeper) to work with Sesame.
It has the following subdirectories:
- **policies/** has sample `ConstraintTemplates` and `Constraints` implementing rules that a Sesame user *may* want to enforce for their clusters, but that are not required for Sesame to function. You should take a pick-and-choose approach to the contents of this directory, and should modify/extend them to meet your unique needs.
- **validations/** has `ConstraintTemplates` and `Constraints` implementing rules that Sesame universally requires to be true. If you're using Sesame and Gatekeeper, we recommend you use all of the rules defined in this directory.

See the [Gatekeeper guide](https://projectsesame.io/guides/gatekeeper/) for more information.
