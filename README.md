![](https://raw.githubusercontent.com/appsody/website/master/src/images/appsody_full_logo.svg?sanitize=true)

## Welcome to Appsody!
<https://appsody.dev>

#### Compose a cloud native masterpiece.

Infused with cloud native capabilities from the moment you start, Appsody provides everything you need to iteratively develop applications, ready for deployment to Kubernetes environments. Teams are empowered with sharable technology stacks, configurable and controllable through a central hub.

# appsody-controller 
The appsody-controller is an appsody process that runs inside the docker container, it does any preinstall required, starts the server with the command specified for run, debug or test mode, and watches for file changes, if configured to do so.  If so configured, when file or directory changes occur it runs a an "ON_CHANGE" action specific to the mode.

## Build Notes: 

### Travis Build
The project is instrumented with Travis CI and with an appropriate Makefile. Most of the build logic is triggered from within the Makefile.

Upon commit, only the test and lint actions are executed by Travis.

In order for Travis to go all the way to package and deploy, you need to create a new release (one that is tagged with a never seen before tag). When you create a new release, a Travis build with automatically run, and the resulting artifacts will be posted on the Releases page.

###Manual Build
You can also test the build process manually.

After setting the GOPATH env var correctly, just run make <action...> from the command line, within the same directory where Makefile resides. For example make package clean will run the package and then the clean actions.

example make commands:  
  
make lint  - Lints the files   
make test - Automated test cases are run
make build - Builds binary in build dir of this folder  
make package - Builds the linux binary and stores it in the package/ dir  

## Where to find the latest binary build

[https://github.com/appsody/controller/releases] 

The controller needs to be executable.  Note that you may have to chmod the permissions to be executable. 

## Environment Variables 
To use the controller, you must configure the relevant Dockerfile environment variables. For more information about the environment variables currently supported by Appsody, please see [here](https://appsody.dev/docs/stacks/environment-variables).

## Running the controller for development:  

Run the appsody-controller-vlatest (appsody-controller) command located in the build directory 

There are two parameters -mode (--mode) and -debug (--debug)

__-mode__ controls which mode the controller runs in: run, debug or test.
If the mode option is unspecified, the mode is assumed to be "run".

The options are as follows (note --mode and -mode are interchangeable):  
>No value      Mode is run  
>--mode      Mode is run  
>--mode=debug    Mode is debug  
>--mode=test     Mode is test  
>--mode=run      Mode is run  

__-v or -verbose__ controls whether debug level of logging is enabled.
If the mode is not specified it is assume to be false.

This can be specified in several ways (note --verbose and -v are interchangeable):  
>No value        Debug logging is off  
>--v      Debug logging is on  
>--v=false   Debug logging is off  
>--v=true   Debug logging is on  

__--version returns the current version

## The docker appsody/init-controller:{travis_tag} image

This image is built as part of the release/deploy process in travis.
This image contains a copy of the appsody-controller binary placed at the `/` directory of the image.

In addition there is a CMD in the Dockerfile build which will cause the /appsody-controller binary to be copied to the /.appsody directory, which is a volume known to the Appsody CLI.

`CMD ["cp","/appsody-controller","/.appsody/appsody-controller"]`

## Building a test init-controller docker image

The developer can build there own init-controller image to test with the CLI by usin the build.sh script.

Specify the following environment variables:
- TRAVIS_TAG 
- DOCKER_ORG
- DOCKER_USERNAME
- DOCKER_PASSWORD

This will build the init-controller image locally and push it to the corresponding DOCKER_ORG/DOCKER_USERNAME/init-controller:{TRAVIS_TAG} docker repository location where it will also be tagged as 'latest'.


## Testing the controller with the Appsody CLI
The Appsody CLI makefile specifies a particular version of the appsody-controller.  Which will cause a particular version of init-controller to be used.

If a tester needs to test a different version of the controller with the CLI, there are two environment variables they can use to test with a different version:

APPSODY_CONTROLLER_VERSION will allow the tester to specify a different version of the CLI for instance 0.3.0 vs 0.3.1
APPSODY_CONTROLLER_IMAGE will allow the test to specify a different image to use, this is very useful during development.
For instance a developer could create a new image for the controller at docker.io/{org}/init-controller:{tag}

## Controller behavior

- As of release 0.2.4 only file related events will trigger ON_CHANGE actions by the controller.  Directory events such as mkdir, rmdir, chmod, etc will not trigger ON_CHANGE actions.
- As of release 0.2.4 a potential problem with how events occuring in the APPSODY_WATCH_IGNORE_DIR are handled has been fixed.  Such events are now pre processed by the watcher code, rather than post processed once the event reaches the controller.

## Known issues

If the appsody stack of interest uses a script file (.sh for example) that is then edited by the `vi` editor while the script is running, the file modification time is not updated on the container file system until the script ends.  What this means is that the ON_CHANGE action is not triggered when `vi` writes the file.

## Contributing

We welcome all contributions to the Appsody project. Please see our [Contributing guidelines](https://github.com/appsody/docs/blob/master/CONTRIBUTING.md)

## Community

Join our [Slack](https://appsody-slack.eu-gb.mybluemix.net/) to meet the team, ask for help, and talk about Appsody!

## Licence

Please see [LICENSE](https://github.com/appsody/docs/blob/master/LICENSE) and [NOTICES](https://github.com/appsody/website/blob/master/NOTICE.md) for more information.
