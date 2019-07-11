# How to make this asset available

The Appsody Controller is made available by creating a tagged GitHub release:
* Go to the _Releases_ page of the repo
* Click _Draft a new release_
* Define a tag in the format of x.y.z (example: 0.2.1). Use the tag also for the title.
* Describe the release with your release notes, including a list of the features added by the release, a list of the major issues that are resolved by the release, caveats, known issues.
* Click _Publish release_

These steps will trigger the `deploy` stage of the build process, as defined in `.travis.yml`. The `deploy` stage, if successful, will produce the following results:
* The release page will be populated with the build artifacts (binary for the controller, Source Code .zip and Source Code .tar.gz files)

# Release schedule
We plan to release the Appsody Controller at the end of each sprint - approximately every two weeks.

# Dependencies
Currently the Appsody Controller has no dependencies
## Downstream Dependencies
The Appsody CLI is heavily dependant on Appsody Controller.  When a new Appsody Controller release is created, a corresponding Appsody CLI release must be created.  See the release notes process for Appsody CLI [here](https://github.com/appsody/appsody/blob/master/RELEASE.md)

