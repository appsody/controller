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
To use the controller, the following Dockerfile environment variables must be configured:

Note: values given as illustrations


### APPSODY_WATCH_INTERVAL (Optional)

This is the watch interval (in seconds).  This is optional.  The default is 2 seconds.

>APPSODY_WATCH_INTERVAL=3


### APPSODY_MOUNTS:  (Potentially optional)
This variable contains the mount directories, which can alternatively be used as watch directories if no value exists for APPSODY_WATCH_DIR.  There can be multiple mount directories separated by a';'.  The format is <localdir>:/<docker container directory>.

>ENV APPSODY_MOUNTS=/:/project/user-app

### APPSODY_WATCH_DIR: (Potentially optional)

This variable contains the watch directories to watch for changes in. There can be multple directories separated by a ';'.  The format is dir1;dir2.

The value of APPSODY_MOUNTS can be used in place of APPSODY_WATCH_DIR.

>ENV APPSODY_WATCH_DIR=/project/user-app

### APPSODY_WATCH_IGNORE_DIR
This variable contains the directories to ignore any changes in. There can be multiple directories separated by a ';'.  The format is dir1;dir2.

>ENV APPSODY_WATCH_IGNORE_DIR=/project/user-app/node_modules

### APPSODY_WATCH_REGEX:  
This is a regex expression which describes which files are watched for changes.  Currently negative look ahead matching (e.g. ignore patterns) is not supported.  If this value is not supplied, it will default to watch for changes in .java, .go, and .js files. 

>ENV APPSODY_WATCH_REGEX="(^.*.java$)|(^.*.js$)|(^.*.go$)"

### APPSODY_PREP

This is an optional command executed before the APPSODY_RUN/TEST/DEBUG and APPSODY_RUN/TEST/DEBUG_ON_CHANGE commands are run. This command should only be used to perform prerequisite checks or preparation steps prior to starting the app server. If this command fails, APPSODY_RUN/TEST/DEBUG will not be executed and the appsody container will be terminated. It is _not_ recommended to perform code compilation tasks in APPSODY_PREP because compilation errors can typically be fixed and recovered while the container is running with the APPSODY_RUN/TEST/DEBUG and ON_CHANGE commands. Unlike those commands, APPSODY_PREP will only be run once and never retried.

__Note:__ APPSODY_INSTALL is deprecated and has been replaced with APPSODY_PREP

>ENV APPSODY_PREP="npm install --prefix user-app"

### APPSODY_RUN

This is the command run for the server process after the APPSODY_PREP command, when the mode is 'run'.  
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script. 

>ENV APPSODY_RUN="npm start"  

### APPSODY_RUN_ON_CHANGE

This is the command run when a change is detected on the file system by the controller when the mode is 'run'.
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script.   
If the file watching is disabled, the value should be "".
  
>ENV APPSODY_RUN_ON_CHANGE="npm start" 

### APPSODY_RUN_KILL

APPSODY_RUN_KILL is used to signal that when the mode is "run" the controller will kill the server process started by APPSODY_RUN prior 
to starting the watch action specified by APPSODY_RUN_ON_CHANGE.  The values supported are true or false.  The default is "true".

>ENV APPSODY_RUN_KILL=<true/false>

### APPSODY_DEBUG

This is the command for the server process run after the APPSODY_PREP command, when the mode is 'debug'.  
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script. 

>ENV APPSODY_DEBUG="npm run debug"

### APPSODY_DEBUG_ON_CHANGE

This is the command run when a change is detected on the file system by the controller when the mode is 'debug'.
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script.   
If the file watching is disabled, the value should be "".
  
>ENV APPSODY_DEBUG_ON_CHANGE="npm run debug"

### APPSODY_DEBUG_KILL 
This variable isused to signal that when the mode is "debug" the controller will kill the server process started by APPSODY_DEBUG prior to starting the watch action specified by APPSODY_DEBUG_ON_CHANGE.  The values supported are true or false.  The default is "true".

>APPSODY_DEBUG_KILL=<true/false>


### APPSODY_TEST

This is the command to run the test cases run after the APPSODY_PREP command, when the mode is 'test'.  
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script. 

>ENV APPSODY_TEST="npm test && npm test --prefix user-app"

### APPSODY_TEST_ON_CHANGE

This is the command run when a change is detected on the file system by the controller when the mode is 'test'.
If your command involves complex environment variable expansions, it may be better to encapsulate your command into a script.   
If the file watching is disabled, the value should be "".

>ENV APPSODY_TEST_ON_CHANGE=""


### APPSODY_TEST_KILL 

This variable is used to signal that when the mode is "test" the controller will kill the server process started by APPSODY_TEST prior to starting the watch action specified by APPSODY_TEST_ON_CHANGE.  The values supported are true or false.  The default is "true".

>APPSODY_TEST_KILL=<true/false>

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

## Contributing

We welcome all contributions to the Appsody project. Please see our [Contributing guidelines](https://github.com/appsody/docs/blob/master/CONTRIBUTING.md)

## Community

Join our [Slack](https://appsody-slack.eu-gb.mybluemix.net/) to meet the team, ask for help, and talk about Appsody!

## Licence

Please see [LICENSE](https://github.com/appsody/docs/blob/master/LICENSE) and [NOTICES](https://github.com/appsody/website/blob/master/NOTICE.md) for more information.