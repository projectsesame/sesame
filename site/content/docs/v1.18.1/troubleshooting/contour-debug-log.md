# Enabling Sesame Debug Logging

The `Sesame serve` subcommand has two command-line flags that can be helpful for debugging.
The `--debug` flag enables general Sesame debug logging, which logs more information about how Sesame is processing API resources.
The `--kubernetes-debug` flag enables verbose logging in the Kubernetes client API, which can help debug interactions between Sesame and the Kubernetes API server.
This flag requires an integer log level argument, where higher number indicates more detailed logging.
