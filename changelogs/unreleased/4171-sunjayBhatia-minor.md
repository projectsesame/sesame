### Leader Election Configuration

`Sesame serve` leader election configuration via config file has been deprecated.
The preferred way to configure leader election parameters is now via command line flags.
See [here](https://projectsesame.io/docs/v1.20.0/configuration/#serve-flags) for more detail on the new leader election flags.

*Note:* If you are using the v1alpha1 SesameConfiguration CRD, leader election configuration has been removed from that CRD as well.
Leader election configuration is not something that will be dynamically configurable once Sesame implements configuration reloading via that CRD.
