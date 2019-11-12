# How to make this asset available

The Appsody Controller is made available by creating a tagged GitHub release:
* Go to the _Releases_ page of the repo
* Click _Draft a new release_
* Define a tag in the format of x.y.z (example: 0.2.1). Use the tag also for the title.
* Describe the release with your release notes, including a list of the features added by the release, a list of the major issues that are resolved by the release, caveats, known issues.
* Click _Publish release_

### Monitor the build
1. Watch the [Travis build](https://travis-ci.com/appsody/controller) for the release and ensure it passes. The build will include the `deploy` stage of the build process as defined in `.travis.yml`. The `deploy` stage, if successful, will produce the following results:
    * The release page will be populated with the build artifacts:
    * appsody-controller
    * Source Code .zip
    * Source Code .tar.gz files
    * The image `appsody/init-controller:{tag}` will be pushed to docker.io
        - Where `tag` is the travis tag of the build in Travis CI
        - Note, the init-controller image will also be tagged as `latest`
    * Check to make sure this image has been created
   
# Release schedule
Appsody Controller is released as needed.

# Dependencies
Currently the Appsody Controller has no dependencies
## Downstream Dependencies
The Appsody CLI is heavily dependant on Appsody Controller.  

In the current design, the CLI pulls the dockier.io/appsody/init-controller:{tag} image as needed. Subsequently, the appsody-controller binary is copied to a volume as needed.

When a new Appsody Controller release is created, a corresponding Appsody CLI release must be created.  See the release notes process for Appsody CLI [here](https://github.com/appsody/appsody/blob/master/RELEASE.md)

Specifically the appsody/appsody Makefile variable APPSODY_CONTROLLER_VERSION must be updated for the CLI when new controller versions are used.



