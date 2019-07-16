package main

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

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
	"k8s.io/klog"
)

var (
	VERSION = "vlatest"
)
var appsodyWATCHIGNOREDIR []string
var appsodyWATCHDIRS []string    // # regex of dirs/files to watch for changes. optional, default to mounts
var appsodyRUNWATCHACTION string //# command to run when files change, optional, in java only need to recompile not start server
var appsodyDEBUG string
var appsodyRUN string
var appsodyTEST string
var appsodyINSTALL string
var appsodyMOUNTS []string

var appsodyWATCHREGEX string
var appsodyPREP string
var appsodyWATCHINTERVAL time.Duration
var appsodyDEBUGWATCHACTION string
var appsodyTESTWATCHACTION string
var appsodyRUNKILL bool
var appsodyDEBUGKILL bool
var appsodyTESTKILL bool
var workDir string
var klogFlags *flag.FlagSet
var verbose bool
var vmode bool

type ProcessType int

const (
	server      ProcessType = 0
	fileWatcher ProcessType = 1
)

type controllerManagedProcesses struct {
	pids      map[ProcessType]int
	processes map[ProcessType]*os.Process
	mu        sync.RWMutex
}

var (
	cmps *controllerManagedProcesses
	once sync.Once
)

func appsodyControllerManagedProcesses() *controllerManagedProcesses {
	once.Do(func() {
		cmps = &controllerManagedProcesses{
			pids:      make(map[ProcessType]int),
			processes: make(map[ProcessType]*os.Process),
		}
	})

	return cmps
}

type envError struct {
	environmentVar1 string
	environmentVar2 string
	environmentVar3 string
}

func (e envError) Error() string {

	errorReturn := fmt.Sprintf("%v and %v and %v can not be empty.", e.environmentVar1, e.environmentVar2, e.environmentVar3)

	return errorReturn

}

type volumesError struct {
	environmentVar1 string
	environmentVar2 string
}

func (e volumesError) Error() string {

	errorReturn := fmt.Sprintf("%v and %v can not be empty. File watching is enabled.", e.environmentVar1, e.environmentVar2)

	return errorReturn

}

type appsodylogger string

func (l appsodylogger) log(args ...interface{}) {
	var dbg [1]interface{}

	if l != "Info" {

		dbg[0] = "[" + l + "] "
		args = append(dbg[0:], args...)
	}

	if l == "Debug" {
		if klog.V(2) {

			// we don't want to pring out debug unless debug level is set
			klog.InfoDepth(1, args...)
		}
	} else {

		klog.InfoDepth(1, args...)
	}

}

var (

	// Info - informational logging
	Info appsodylogger = "Info"
	// Warning - warning logging
	Warning appsodylogger = "Warning"
	// Error - error logging
	Error appsodylogger = "Error"
	// Fatal - fatal errors
	Fatal appsodylogger = "Fatal"
	// Debug - debug
	Debug appsodylogger = "Debug"
)

type mountError struct {
	mountsString string
}

func (e mountError) Error() string {
	return fmt.Sprintf("The Mount string has bad formatting: %v", e.mountsString)
}
func computeSigInt(tempSigInt string) bool {
	var answer bool
	if tempSigInt == "" || strings.Compare(strings.TrimSpace(strings.ToUpper(tempSigInt)), "TRUE") == 0 {

		answer = true
	} else {

		answer = false
	}
	return answer
}
func setupEnvironmentVars() error {

	Debug.log("setupEnvironmentVars ENTRY")
	var err error

	tmpWATCHIGNOREDIR := os.Getenv("APPSODY_WATCH_IGNORE_DIR")
	Debug.log("APPSODY_WATCH_IGNORE_DIR ", tmpWATCHIGNOREDIR)

	appsodyRUNKILL = computeSigInt(os.Getenv("APPSODY_RUN_KILL"))
	Debug.log("APPSODY_RUN_KILL ", appsodyRUNKILL)
	appsodyDEBUGKILL = computeSigInt(os.Getenv("APPSODY_DEBUG_KILL"))
	Debug.log("APPSODY_DEBUG_KILL ", appsodyRUNKILL)
	appsodyTESTKILL = computeSigInt(os.Getenv("APPSODY_TEST_KILL"))
	Debug.log("APPSODY_DEBUG_KILL ", appsodyRUNKILL)

	appsodyDEBUGWATCHACTION = os.Getenv("APPSODY_DEBUG_ON_CHANGE")
	Debug.log("APPSODY_DEBUG_ON_CHANGE: " + appsodyDEBUGWATCHACTION)

	appsodyTESTWATCHACTION = os.Getenv("APPSODY_TEST_ON_CHANGE")
	Debug.log("APPSODY_TEST_ON_CHANGE: " + appsodyTESTWATCHACTION)

	appsodyTEST = os.Getenv("APPSODY_TEST")
	Debug.log("APPSODY_TEST: " + appsodyTEST)

	appsodyWATCHREGEX = os.Getenv("APPSODY_WATCH_REGEX")
	Debug.log("APPSODY_WATCH_REGEX: " + appsodyWATCHREGEX)

	// if there is no watch expression default to watching for .go,.java,.js files
	if appsodyWATCHREGEX == "" {
		appsodyWATCHREGEX = "(^.*.java$)|(^.*.js$)|(^.*.go$)"
	}

	appsodyRUN = os.Getenv("APPSODY_RUN")
	Debug.log("APPSODY_RUN: " + appsodyRUN)

	tmpWatchDirs := os.Getenv("APPSODY_WATCH_DIR")
	Debug.log("APPSODY_WATCH_DIR: " + tmpWatchDirs)

	appsodyRUNWATCHACTION = os.Getenv("APPSODY_RUN_ON_CHANGE")
	Debug.log("APPSODY_RUN_ON_CHANGE: " + appsodyRUNWATCHACTION)

	appsodyINSTALL = os.Getenv("APPSODY_INSTALL")
	Debug.log("APPSODY_INSTALL: " + appsodyINSTALL)

	appsodyPREP = os.Getenv("APPSODY_PREP")
	Debug.log("APPSODY_PREP: " + appsodyPREP)
	if appsodyPREP == "" {
		appsodyPREP = appsodyINSTALL
	}

	appsodyDEBUG = os.Getenv("APPSODY_DEBUG")
	Debug.log("APPSODY_DEBUG: " + appsodyDEBUG)

	tmpMountDirs := os.Getenv("APPSODY_MOUNTS")
	Debug.log("APPSODY_MOUNTS: " + tmpMountDirs)

	tempWatchInterval := os.Getenv("APPSODY_WATCH_INTERVAL")
	Debug.log("APPSODY_WATCH_INTERVAL: " + tempWatchInterval + " seconds")
	var value int
	var atoiErr error
	if tempWatchInterval != "" {
		trimmedInterval := strings.TrimSpace(tempWatchInterval)
		value, atoiErr = strconv.Atoi(trimmedInterval)

		if atoiErr != nil {

			Warning.log("Invalid watch interval, setting to default 2000: " + tempWatchInterval)

			value = 2
		}

	} else {
		// default to 2 seconds
		value = 2
	}

	appsodyWATCHINTERVAL = time.Duration(int64(value) * int64(time.Second))

	Debug.log("appsodyWATCHINTERVAL set to: ", appsodyWATCHINTERVAL)
	fileWatchingOff := false
	if appsodyRUNWATCHACTION == "" && appsodyDEBUGWATCHACTION == "" && appsodyTESTWATCHACTION == "" {
		Debug.log("File watching is off.")
		fileWatchingOff = true
	}

	if appsodyDEBUG == "" && appsodyRUN == "" && appsodyTEST == "" {
		err = envError{"APPSODY_DEBUG", "APPSODY_RUN", "APPSODY_TEST"}
		return err
	} else if !fileWatchingOff && tmpMountDirs == "" && tmpWatchDirs == "" {
		err = volumesError{"APPSODY_WATCH_DIR", "APPSODY_MOUNTS"}
		return err

	}

	// split the watch dirs using ; separator
	if tmpWatchDirs != "" {

		appsodyWATCHDIRS = strings.Split(tmpWatchDirs, ";")
		for i := 0; i < len(appsodyWATCHDIRS); i++ {
			appsodyWATCHDIRS[i] = strings.TrimSpace(appsodyWATCHDIRS[i])

		}

	}
	if tmpWATCHIGNOREDIR != "" {

		appsodyWATCHIGNOREDIR = strings.Split(tmpWATCHIGNOREDIR, ";")
		for i := 0; i < len(appsodyWATCHIGNOREDIR); i++ {
			appsodyWATCHIGNOREDIR[i] = strings.TrimSpace(appsodyWATCHIGNOREDIR[i])

		}

	}

	// split the mount dirs using ; separator
	if tmpMountDirs != "" {

		appsodyMOUNTS = strings.Split(tmpMountDirs, ";")
		for i := 0; i < len(appsodyMOUNTS); i++ {
			// check if there is a : separator
			if strings.Contains(appsodyMOUNTS[i], ":") {
				localDir := strings.Split(appsodyMOUNTS[i], ":")
				//Windows may prepend the drive ID to the path so just take the last split
				//ex. C:\whatever\path\:/linux/dir
				appsodyMOUNTS[i] = strings.TrimSpace(localDir[len(localDir)-1])
			} else {
				err = mountError{tmpMountDirs}
				break
			}

		}

	}

	Debug.log("setupEnvironmentVars EXIT")

	return err
}

func killProcess(theProcessType ProcessType) error {
	var processPid int
	var err error
	Debug.log("killProcess Entry")
	if theProcessType == server {
		processPid = cmps.pids[server]
	} else {
		processPid = cmps.pids[theProcessType]
	}
	Debug.log("Process pid is: ", processPid)

	if processPid != 0 {

		Debug.log("Killing pid:  ", processPid)
		err = syscall.Kill(-processPid, syscall.SIGINT)

		cmps.processes[theProcessType] = nil
		cmps.pids[theProcessType] = 0
		if err != nil {

			Error.log("Killing process returned an error SIGINT received error ", err)

		}
	}
	return err
}

/*
	runInstall
*/
func runInstall(commandString string) (*exec.Cmd, error) {
	var err error
	Info.log("Running Install: " + commandString)
	cmd := exec.Command("/bin/bash", "-c", commandString)
	Debug.log("Set workdir:  " + workDir)
	cmd.Dir = workDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Run()
	Debug.log("Install complete")

	return cmd, err
}

/*
	StartProcess
*/
func startProcess(commandString string, theProcessType ProcessType) (*exec.Cmd, error) {
	var err error
	Info.log("Running: " + commandString)
	cmd := exec.Command("/bin/bash", "-c", commandString)
	Debug.log("Set workdir:  " + workDir)
	cmd.Dir = workDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()

	cmps.processes[theProcessType] = cmd.Process

	cmps.pids[theProcessType] = cmd.Process.Pid
	Debug.log("New process created with pid ", strconv.Itoa(cmd.Process.Pid))

	return cmd, err
}

func waitProcess(cmd *exec.Cmd, theProcessType ProcessType) error {
	Debug.log("waitProcess ENTRY")
	err := cmd.Wait()
	if err != nil {

		// do nothing as the kill causees and error condition
		Info.log("Wait received error:", err)

	}

	Debug.log("waitProces EXIT")
	return err
}

func unwatchDir(path string) bool {
	unwatch := false
	if appsodyWATCHIGNOREDIR != nil {
		for i := 0; i < len(appsodyWATCHIGNOREDIR); i++ {
			if strings.HasPrefix(path, appsodyWATCHIGNOREDIR[i]) {
				unwatch = true
				break
			}
		}
	}
	return unwatch

}

func runWatcher(fileChangeCommand string, dirs []string, killServer bool) error {
	errorMessage := ""
	var err error

	Debug.log("Run watcher ENTRY: " + fileChangeCommand)
	// Start the Watcher
	// compile the regex prior to running watcher because panic leaves child processes if it occurs

	r := regexp.MustCompile(appsodyWATCHREGEX)
	w := watcher.New()
	for d := 0; d < len(dirs); d++ {
		// Watch each directory specified recursively for changes.
		currentDir := dirs[d]
		// Make sure the directory exists
		_, err = os.Stat(currentDir)
		if err != nil {

			errorMessage = "Watched directory does not exist: " + currentDir
			Warning.log(errorMessage, err)

		}
		if err = w.AddRecursive(currentDir); err != nil {

			errorMessage = "Failed to add directory to recursive watch list: " + currentDir
			Warning.log(errorMessage, err)

		}

	}

	w.SetMaxEvents(1)

	// Only files that match the regular expression during file listings
	// will be watched.  Currently we watch java, js, and go files.
	// We may add an environment variable to add to this list

	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case event := <-w.Event:
				Debug.log("Full path is:  " + event.Path)
				if unwatchDir(event.Path) {
					Debug.log("Path is not to be watched")
				} else {
					Debug.log("Matching for:  " + event.Name())
					if r.MatchString(event.Name()) {

						Debug.log("About to restart watcher")

						// Restart the watcher as a thread so we can do a wait to avoid zombie in ps -ef
						if fileChangeCommand != "" {
							Debug.log("kill server is: ", killServer)
							go runCommands(fileChangeCommand, fileWatcher, killServer)
						}

					}
				}

			case err := <-w.Error:
				Warning.log("An error occured in the watcher ", err)

			case <-w.Closed:
				Debug.log("Watcher is now closed")
				return
			}
		}
	}()

	// Start the watching process - it'll check for changes every "n" ms.
	Debug.log("Watch interval is: ", appsodyWATCHINTERVAL)
	if err = w.Start(appsodyWATCHINTERVAL); err != nil {
		errorMessage = "Could not start the watcher "
		Error.log(errorMessage+" ", err)
	}
	defer w.Close()
	// Close the watcher at function end
	//defer w.Close()

	Debug.log("Run watcher EXIT: ")
	return err
}

/*
   determine if we need to kill the server process

*/
func runCommands(commandString string, theProcessType ProcessType, killServer bool) {

	var cmd *exec.Cmd
	var err error
	var mutexUnlocked bool

	// Start a new watch action
	Debug.log("ENTRY Running command:  " + commandString)
	cmps := appsodyControllerManagedProcesses()

	// lock the mutex which protects the Process and the Pid String
	cmps.mu.Lock()
	Debug.log("Mutex Locked")
	if theProcessType == server {

		if appsodyPREP != "" {
			_, err = runInstall(appsodyPREP)
		}
		if err != nil {
			Error.log("FATAL error APPSODY_PREP command received an error.  The controller is exiting: ", err)
			os.Exit(1)
		}
		// keep going
		cmd, err = startProcess(commandString, server)
		Debug.log("Started Server process")
		if err != nil {
			Warning.log("ERROR start server (APPSODY_RUN) received error ", err)
		}
		cmps.mu.Unlock()
		mutexUnlocked = true
		Debug.log("mutex unlocked")
		err = waitProcess(cmd, theProcessType)

		if err != nil {
			Info.log("Wait received error on server start ", err)
		}
	} else {
		Debug.log("Inside watcher path")
		// This is a watcher
		if killServer {
			Debug.log("killing server")
			err = killProcess(server)
			if err != nil {
				// do nothing we continue after kill errors
				Warning.log("killProcess received error ", err)
			}
		}
		err = killProcess(fileWatcher)
		if err != nil {
			// do nothing we continue after kill errors
			Warning.log("Watcher killProcess received error ", err)
		}

		commandToUse := commandString
		processTypeToUse := fileWatcher

		if !killServer {
			// this path is only relevant for APPSODY_<RUN/DEBUG/TEST>KILL_SERVER=FALSE
			// get the process of the current server (should not be nil ever) and send benign SIG 0 to the server proces
			if cmps.processes[server] != nil && cmps.processes[server].Signal(syscall.Signal(0)) != nil {
				// if there is no server process, an error is returned
				Debug.log("The server process with pid:", cmps.processes[server].Pid, "was not found, and APPSODY_<action>_KILL is set to false. The server will be restarted.")
				//start the server with the startCommand, not the watch action command
				commandToUse = startCommand
				processTypeToUse = server
			}
		}

		cmd, err = startProcess(commandToUse, processTypeToUse)

		if err != nil {
			Warning.log("ERROR start process received error: ", err)
		}
		cmps.mu.Unlock()
		mutexUnlocked = true
		Debug.log("mutex unlocked")
		err = waitProcess(cmd, processTypeToUse)
		if err != nil {
			// do nothing as the kill causees and error condition
			Info.log("Wait received error ", err)
		}

	}

	if !mutexUnlocked {

		cmps.mu.Unlock()
	}

	Debug.log("runCommands EXIT")

}

var startCommand string

func main() {

	var err error
	var fileChangeCommand string
	debugMode := false
	testMode := false
	var dirs []string
	var stopWatchServerOnChange bool

	errorMessage := ""
	var errWorkDir error

	mode := flag.String("mode", "run", "This is the mode the controller runs in: run, debug or test")
	flag.BoolVar(&verbose, "verbose", false, "Turns on debug output and logging ")
	flag.BoolVar(&vmode, "v", false, "Turns on debug output and logging ")
	flag.Parse()
	klogFlags = flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	if vmode || verbose {
		// set debug mode

		_ = klogFlags.Set("v", "4")
		_ = klogFlags.Set("skip_headers", "false")

	} else {
		_ = klogFlags.Set("skip_headers", "true")
	}

	Debug.log("Controller main ENTRY " + VERSION)

	if strings.Compare(*mode, "test") == 0 {
		testMode = true
	}

	if strings.Compare(*mode, "debug") == 0 {
		debugMode = true
	}

	workDir, errWorkDir = os.Getwd()

	if errWorkDir != nil {

		Fatal.log("Could not find the working dir ", errWorkDir)
		os.Exit(1)
	}
	// Obtain the environment variables
	err = setupEnvironmentVars()
	if err != nil {
		errorMessage = "Warning: setup did not find all environment variables "
		Fatal.log(errorMessage, err)
		os.Exit(1)
	}

	// Set the startCommand based upon whether debug Mode is enabled
	if debugMode {
		startCommand = appsodyDEBUG
	} else if testMode {
		startCommand = appsodyTEST
	} else {
		startCommand = appsodyRUN
	}
	if startCommand == "" {
		Warning.log("Warning: startCommand (environment variable APPSODY_DEBUG,APPSODY_TEST or APPSODY_RUN) is unspecified")
	}
	Debug.log("startCommand: " + startCommand)
	if debugMode {
		//note this could be ""
		fileChangeCommand = appsodyDEBUGWATCHACTION
	} else if testMode {
		fileChangeCommand = appsodyTESTWATCHACTION
	} else {
		fileChangeCommand = appsodyRUNWATCHACTION
	}
	Debug.log("File change command: " + fileChangeCommand)

	if fileChangeCommand == "" {
		Debug.log("fileChangeCommand environment variable APPSODY_WATCH_ACTION) is unspecified.")
		Debug.log("Running sync: " + startCommand)
		runCommands(startCommand, server, false)
	} else {
		Debug.log("Running " + startCommand)
		go runCommands(startCommand, server, false)
	}

	// use the appropriate server on change setting
	if debugMode {

		stopWatchServerOnChange = appsodyDEBUGKILL
	} else if testMode {
		stopWatchServerOnChange = appsodyTESTKILL
	} else {
		stopWatchServerOnChange = appsodyRUNKILL
	}

	// Prefer the watch dirs be set to the APPSODY_WATCH_DIR value, but fall back to the APPSODY_MOUNTS if need be

	if appsodyWATCHDIRS != nil {
		dirs = appsodyWATCHDIRS
	} else {
		dirs = appsodyMOUNTS
	}

	if fileChangeCommand != "" {

		err = runWatcher(fileChangeCommand, dirs, stopWatchServerOnChange)
	} else {
		Info.log("Watcher is not running.")
	}
	if err != nil {
		errorMessage = "Error running watcher "
		Fatal.log(errorMessage, err)
		os.Exit(1)
	}

	Debug.log("Controller main EXIT")
}
