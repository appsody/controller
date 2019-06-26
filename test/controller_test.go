// Copyright © 2019 IBM Corporation and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package test

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

// TestInvalidWatchDirs
// The Watch directories can not be unspecified
func TestInvalidWatchDirs(t *testing.T) {
	log.Println("TestInvalidWatchDirs")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("Setup Path", func(t *testing.T) {
		/*
			The following environment variables are unset
			APPSODY_WATCH_DIR
			APPSODY_MOUNTS
			The following environment variables are set
			APPSODY_RUN
			APPSODY_DEBUG
			APPSODY_RUN_ON_CHANGE
			The controller is invoked with verbose logging

			The test checks to see if empty APPSODY_WATCH_DIR and APPSODY_MOUNTS are flagged.
		*/
		args := []string{"unset APPSODY_WATCH_DIR;unset APPSODY_MOUNTS;export APPSODY_RUN=\"echo run\";export APPSODY_DEBUG=\"echo debug \";echo $APPSODY_RUN;echo $APPSODY_DEBUG;export APPSODY_RUN_ON_CHANGE=\"sleep 2\" ;go run ../main.go  -v=true"}

		output, err := RunBashCmdExec(args, ".")
		log.Println(output)

		if err != nil && strings.Contains(output, "APPSODY_WATCH_DIR and APPSODY_MOUNTS can not be empty") {
			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

// TestWatchAction
// The only valid input params to main are: debug run <none>, if there is more than one it is an error
func TestWatchAction(t *testing.T) {
	log.Println("TestWatchAction")
	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestWatchAction", func(t *testing.T) {
		/*
			The Following environment variables are set:
			APPSODY_WATCH_EXPRESS - to match java files
			APPSODY_WATCH_DIR
			APPSODY_WATCH_INTERVAL
			APPSODY_MOUNTS
			APPSODY_RUN_ON_CHANGE
			APPSODY_RUN
			APPSODY_DEBUG

			The controller is invoked with verbose logging.
			The output is checked for the correct watch interval, APPSODY_ON_CHANGE command and APPSODY_RUN command
		*/
		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_WATCH_INTERVAL=1;export APPSODY_MOUNTS=\"c:\\bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"sleep 2\" ;export APPSODY_RUN=\"sleep 10\";export APPSODY_DEBUG=\"echo debug \";env |grep APPSODY;go run ../main.go -v=true"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)
		// the count should only be one, it will not match on the .go file or the .bad file touched by the util
		// sleep 2 is for watch action, sleep 10 is for RUN action

		if strings.Contains(output, "appsodyWATCHINTERVAL set to: 1s") && strings.Contains(output, "Running: sleep 10") && strings.Count(output, "Running: sleep 2") == 1 {
			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

func TestBadInstall(t *testing.T) {
	log.Println("TestBadInstall")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestBadInstall", func(t *testing.T) {
		/* The following environment variables are set:
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_RUN
		APPSODY_INSTALL -bad command given
		APPSODY_DEBUG
		The controller is executed with verbose logging
		The output is checked for an APPSODY_INSTALL error
		*/
		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_WATCH_INTERVAL=1;export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"sleep 2\";export APPSODY_INSTALL=\"badbad\" ; export APPSODY_RUN=\"sleep 10\";export APPSODY_DEBUG=\"echo debug \";env |grep APPSODY;go run ../main.go -v=true"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)

		if strings.Contains(output, "ERROR Install (APPSODY_INSTALL) received error") {

			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

func TestBadAPPSODYRun(t *testing.T) {
	log.Println("TestBadAPPSODYRun")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestBadAPPSODYRun", func(t *testing.T) {

		/* The following environment variables are set:
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_RUN - bad command given
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed with verbose logging
		The output is checked for a error for server start
		*/
		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_WATCH_INTERVAL=1;export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"sleep 2\";export APPSODY_INSTALL=\"ls\" ; export APPSODY_RUN=\"bad\";export APPSODY_DEBUG=\"echo debug \";env |grep APPSODY;go run ../main.go -v=true"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)

		if strings.Contains(output, "Wait received error on server start exit status") {

			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

func TestBadOnChange(t *testing.T) {
	log.Println("TestBadOnChange")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestBadOnChange", func(t *testing.T) {

		/* The following environment variables are set:
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE -bad value given
		APPSODY_RUN
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed with verbose logging
		The output is checked for a error for the wait for the on change command.
		*/
		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_WATCH_INTERVAL=1;export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"bad\";export APPSODY_INSTALL=\"ls\" ; export APPSODY_RUN=\"ls\";export APPSODY_DEBUG=\"echo debug \";env |grep APPSODY;go run ../main.go -v=true"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)

		if strings.Contains(output, "Wait received error:exit status") {

			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}
func TestBadWatchDir(t *testing.T) {
	log.Println("TestBadWatchDir")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestBadWatchDir", func(t *testing.T) {

		/* The following environment variables are set:
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR - bad value is given
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_RUN
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed with verbose logging
		The output is checked for a error because the watch directory does not exist.
		*/
		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=/tmp/watchdir4;export APPSODY_WATCH_INTERVAL=1;export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"sleep 2\";export APPSODY_INSTALL=\"ls\" ; export APPSODY_RUN=\"ls\";export APPSODY_DEBUG=\"echo debug \";env |grep APPSODY;go run ../main.go -v=true"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)
		// the count should only be one, it will not match on the .go file or the .bad file touched by the util
		// sleep 2 is for watch action, sleep 10 is for RUN action

		if strings.Contains(output, "Watched directory does not exist") {

			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

// TestWatchActionDebug
// The only valid input params to main are: debug run <none>, if there is more than one it is an error
func TestWatchActionDebug(t *testing.T) {
	log.Println("TestWatchActionDebug")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestWatchActionDebug", func(t *testing.T) {

		/* The following environment variables are set:
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_DEBUG_ON_CHANGE - the test checks this value
		APPSODY_RUN
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed in debug mode with verbose logging
		The output is checked the appropriate debug and APPSODY_DEBUG_ON_CHANGE output
		*/

		args := []string{"export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_DEBUG_ON_CHANGE=\"sleep 25\";export APPSODY_RUN_ON_CHANGE=\"sleep 25\" ;export APPSODY_RUN=\"sleep 10\";export APPSODY_DEBUG=\"sleep 10\";env |grep APPSODY;go run ../main.go -v=true -mode=debug"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)
		// the count should only be one, it will not match on the .go file or the .bad file touched by the util
		// sleep 2 is for watch action, sleep 10 is for RUN action
		log.Printf("COUNT OF JAVA TOUCHES: %v %v\n", strings.Contains(output, "Running: sleep 10"), strings.Count(output, "Running: sleep 25"))

		if strings.Contains(output, "Running: sleep 10") && strings.Count(output, "Running: sleep 25") == 1 {
			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

// TestTAA
// The only valid input params to main are: debug run <none>, if there is more than one it is an error
func TestAA(t *testing.T) {
	log.Println("TestAA")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestAA", func(t *testing.T) {

		/* The following environment variables are set:
		APPSODY_TEST - the test checks for this value
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_DEBUG_ON_CHANGE
		APPSODY_RUN
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed in debug mode with verbose logging
		The output is checked the appropriate APPSODY_TEST output
		*/

		args := []string{"export APPSODY_TEST=\"ls -l;sleep 10\";export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_RUN_ON_CHANGE=\"sleep 2\" ;export APPSODY_RUN=\"sleep 10\";export APPSODY_DEBUG=\"sleep 10\";env |grep APPSODY;go run ../main.go -v=true -mode=test"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", false, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)
		// the count should only be one, it will not match on the .go file or the .bad file touched by the util
		// sleep 2 is for watch action, sleep 10 is for RUN action
		if strings.Count(output, "Running: ls -l") == 1 {
			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}

// TestTAAFail
// The only valid input params to main are: debug run <none>, if there is more than one it is an error
func TestTAAFail(t *testing.T) {
	log.Println("TestTAAFail")

	projectDir, err := ioutil.TempDir("", "watchdir2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(projectDir)
	// call t.Run so that we can name and report on individual tests
	t.Run("TestWatchAction", func(t *testing.T) {
		/* The following environment variables are set:
		APPSODY_TEST - a bad value is used
		APPSODY_WATCH_REGEX - watch java files
		APPSODY_WATCH_DIR
		APPSODY_WATCH_INTERVAL
		APPSODY_MOUNTS
		APPSODY_RUN_ON_CHANGE
		APPSODY_DEBUG_ON_CHANGE
		APPSODY_RUN
		APPSODY_INSTALL
		APPSODY_DEBUG
		The controller is executed in debug mode with verbose logging
		The output is checked the appropriate output for the bad test command given for APPSODY_TEST
		*/
		args := []string{"export APPSODY_TEST=\"bad 4\";export APPSODY_WATCH_REGEX=\"(^.*.java$)\";export APPSODY_WATCH_DIR=" + projectDir + ";export APPSODY_MOUNTS=\"/bad:" + projectDir + "\";export APPSODY_TEST_ON_CHANGE=\"sleep 2\" ;export APPSODY_RUN=\"sleep 10\";export APPSODY_DEBUG=\"sleep 10\";env |grep APPSODY;go run ../main.go -v=true -mode=test"}

		output, err := RunBashCmdExecAndKillAndTouch(args, ".", true, projectDir)
		log.Println("This is the output: " + output)

		log.Printf("error is:  %v\n", err)
		// the count should only be one, it will not match on the .go file or the .bad file touched by the util
		// sleep 2 is for watch action, sleep 10 is for RUN action
		if strings.Count(output, "bad: command not found") == 1 {
			log.Println("pass")
		} else {
			t.Fail()
		}

	})

}
