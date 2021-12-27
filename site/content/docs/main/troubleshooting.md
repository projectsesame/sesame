## Troubleshooting

If you encounter issues, follow the guides below for help. For topics not covered here, you can [file an issue][0], or talk to us on the [#Sesame channel][1] on Kubernetes Slack.

### [Envoy Administration Access][2]
Review the linked steps to learn how to access the administration interface for your Envoy instance.

### [Sesame Debug Logging][3]
Learn how to enable debug logging to diagnose issues between Sesame and the Kubernetes API.

### [Envoy Debug Logging][4]
Learn how to enable debug logging to diagnose TLS connection issues.

### [Visualize the Sesame Graph][5]
Learn how to visualize Sesame's internal object graph in [DOT][9] format, or as a png file.

### [Show Sesame xDS Resources][6]
Review the linked steps to view the [xDS][10] resource data exchanged by Sesame and Envoy.

### [Profiling Sesame][7]
Learn how to profile Sesame by using [net/http/pprof][11] handlers. 

### [Sesame Operator][8]
Follow the linked guide to learn how to troubleshoot issues with [Sesame Operator][12].

[0]: {{< param github_url >}}/issues
[1]: {{< param slack_url >}}
[2]: /docs/{{< param latest_version >}}/troubleshooting/envoy-admin-interface/
[3]: /docs/{{< param latest_version >}}/troubleshooting/Sesame-debug-log/
[4]: /docs/{{< param latest_version >}}/troubleshooting/envoy-debug-log/
[5]: /docs/{{< param latest_version >}}/troubleshooting/Sesame-graph/
[6]: /docs/{{< param latest_version >}}/troubleshooting/Sesame-xds-resources/
[7]: /docs/{{< param latest_version >}}/troubleshooting/profiling-Sesame/
[8]: /docs/{{< param latest_version >}}/troubleshooting/operator/
[9]: https://en.wikipedia.org/wiki/Dot
[10]: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol
[11]: https://golang.org/pkg/net/http/pprof/
[12]: https://github.com/projectsesame/sesame-operator
