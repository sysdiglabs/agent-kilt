# agent-kilt

`agent-kilt` provides a convenient way to add additional software to containers by patching ECS TaskDefinitions.



Kilt is a definition on how to include additional software inside containers. It was created to run falco in userspace
along other software. It leverages different runtimes to perform modifications to the container

## Definition file

The kilt definition file contains instructions on how to change a container to run additional software alongside the 
original entry point/command. [HOCON](https://github.com/lightbend/config/blob/master/HOCON.md) was chosen as the 
configuration format.

There are 2 phases of patching that the definition file covers, build time and run time. Build time is when container 
is not running yet. Run time is well, you get it.

### Variables

* **original.*** - contains information about the original container. See runtime specific documentation for details.
    * **original.entry_point** `str`
    * **original.command** `str`
* **build.entry_point** `List[str]` - new entry point
* **build.command** `List[str]` - new command
* **build.environment_variables** `Dict[str,str]` - will merge environment variables
* **build.mount** - add a filesystem inside the target container. Implementation depends on runtime.
    * **build.mount.name** `str` - Mount name
    * **build.mount.image** `str` - the image that contains the volume of the mount
    * **build.mount.volumes** `List(str)` - List of paths to be mounted on the target image
    * **build.mount.entry_point** `List(str)` - The entry point of the image (needed for patching runtimes)
### Example
```
build {
    # concatenated arrays
    entry_point: ["/falco/bin/launcher", "/falco/bin/pdig"] ${?original.entry_point} ${?original.command} ["--"]
    command: ["/falco/usr/bin/falco", "-u", "-c", "/falco/falco.yaml", "--alternate-lua-dir", "/falco/share/lua"]
    environment_variables: {
        "__CW_LOG_GROUP": "FalcoAlerts"
    }
    mount: [
        {
            name: "FalcoDistribution"
            image: "admiral0/falco:latest"
            volumes: ["/falco"]
            entry_point: ["/falco/waitforever"]
        }
    ]
}
```



# Release

To make a new release push a new tag. 

It will trigger the [release-runtime.yml](.github/workflows/release-runtimes.yml) workflow.

For tags, use [semver](https://semver.org/).

Before making a new release, check the diff between the previous tag and `dev` branch,
to spot major changes to make release notes as clear as possible.
