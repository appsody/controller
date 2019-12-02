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
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/appsody/watcher"
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
var appsodyINSTALL string // Note this will be deprecated in a future release
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
var version bool
var interactiveFlag bool
var vmode bool

type ProcessType int

const (
	server      ProcessType = 0
	fileWatcher ProcessType = 1
)

func processTypeToString(theProcessType ProcessType) string {
	if theProcessType == 0 {
		return "APPSODY_RUN/DEBUG/TEST"
	}
	return "APPSODY_RUN/DEBUG/TEST_ON_CHANGE"
}

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

	if l == "ControllerDebug" {
		if klog.V(2) {

			// we don't want to pring out debug unless debug level is set
			klog.InfoDepth(1, args...)
		}
	} else {

		klog.InfoDepth(1, args...)
	}

}

var (

	// ControllerInfo - informational logging
	ControllerInfo appsodylogger = "Info"
	// ControllerWarning - warning logging
	ControllerWarning appsodylogger = "Warning"
	// ControllerError - error logging
	ControllerError appsodylogger = "Error"
	// ControllerFatal - fatal errors
	ControllerFatal appsodylogger = "Fatal"
	// ControllerDebug - debug
	ControllerDebug appsodylogger = "ControllerDebug"
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

	var err error

	tmpWATCHIGNOREDIR := os.Getenv("APPSODY_WATCH_IGNORE_DIR")
	appsodyRUNKILL = computeSigInt(os.Getenv("APPSODY_RUN_KILL"))
	appsodyDEBUGKILL = computeSigInt(os.Getenv("APPSODY_DEBUG_KILL"))
	appsodyTESTKILL = computeSigInt(os.Getenv("APPSODY_TEST_KILL"))
	appsodyDEBUGWATCHACTION = os.Getenv("APPSODY_DEBUG_ON_CHANGE")
	appsodyTESTWATCHACTION = os.Getenv("APPSODY_TEST_ON_CHANGE")
	appsodyTEST = os.Getenv("APPSODY_TEST")
	appsodyWATCHREGEX = os.Getenv("APPSODY_WATCH_REGEX")

	// if there is no watch expression default to watching for .go,.java,.js files
	if appsodyWATCHREGEX == "" {
		appsodyWATCHREGEX = "(^.*.java$)|(^.*.js$)|(^.*.go$)"
	}

	appsodyRUN = os.Getenv("APPSODY_RUN")
	tmpWatchDirs := os.Getenv("APPSODY_WATCH_DIR")
	appsodyRUNWATCHACTION = os.Getenv("APPSODY_RUN_ON_CHANGE")
	appsodyINSTALL = os.Getenv("APPSODY_INSTALL") // Note this will be deprecated in a future release
	appsodyPREP = os.Getenv("APPSODY_PREP")

	if appsodyPREP == "" {
		appsodyPREP = appsodyINSTALL
	}

	appsodyDEBUG = os.Getenv("APPSODY_DEBUG")

	tmpMountDirs := os.Getenv("APPSODY_MOUNTS")

	tempWatchInterval := os.Getenv("APPSODY_WATCH_INTERVAL")

	var value int
	var atoiErr error
	if tempWatchInterval != "" {
		trimmedInterval := strings.TrimSpace(tempWatchInterval)
		value, atoiErr = strconv.Atoi(trimmedInterval)

		if atoiErr != nil {

			ControllerWarning.log("Invalid watch interval, setting to default 2000: " + tempWatchInterval)

			value = 2
		}

	} else {
		// default to 2 seconds
		value = 2
	}

	appsodyWATCHINTERVAL = time.Duration(int64(value) * int64(time.Second))

	fileWatchingOff := false
	if appsodyRUNWATCHACTION == "" && appsodyDEBUGWATCHACTION == "" && appsodyTESTWATCHACTION == "" {
		ControllerDebug.log("File watching is not enabled.")
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

	environmentVars := make(map[string]interface{})

	environmentVars["APPSODY_WATCH_IGNORE_DIR"] = tmpWATCHIGNOREDIR
	environmentVars["APPSODY_DEBUG"] = appsodyDEBUG
	environmentVars["APPSODY_RUN"] = appsodyRUN
	environmentVars["APPSODY_TEST"] = appsodyTEST

	environmentVars["APPSODY_RUN_KILL"] = appsodyRUNKILL
	environmentVars["APPSODY_DEBUG_KILL"] = appsodyDEBUGKILL
	environmentVars["APPSODY_TEST_KILL"] = appsodyTESTKILL
	environmentVars["APPSODY_RUN_ON_CHANGE"] = appsodyRUNWATCHACTION
	environmentVars["APPSODY_DEBUG_ON_CHANGE"] = appsodyDEBUGWATCHACTION
	environmentVars["APPSODY_TEST_ON_CHANGE"] = appsodyTESTWATCHACTION
	environmentVars["APPSODY_WATCH_DIR"] = tmpWatchDirs
	environmentVars["APPSODY_MOUNTS"] = tmpMountDirs
	environmentVars["APPSODY_INSTALL"] = appsodyINSTALL
	environmentVars["APPSODY_PREP"] = appsodyPREP
	environmentVars["APPSODY_WATCH_INTERVAL"] = appsodyWATCHINTERVAL
	environmentVars["APPSODY_WATCH_REGEX"] = appsodyWATCHREGEX
	ControllerDebug.log("Appsody Controller environment variables: ", environmentVars)

	return err
}

func killProcess(theProcessType ProcessType, checkAttempts int) error {
	var processPid int
	var err error
	if theProcessType == server {
		processPid = cmps.pids[server]
	} else {
		processPid = cmps.pids[theProcessType]
	}
	ControllerDebug.log("Attempting to kill pid: ", processPid)

	if processPid != 0 {
		// Check to see if the process is still alive to avoid unncessary kill steps
		if cmps.processes[theProcessType].Signal(syscall.Signal(0)) != nil {
			ControllerDebug.log("No such process for pid:  ", processPid)
			err = nil
		} else {

			ControllerDebug.log("Killing pid:  ", -processPid)
			err = syscall.Kill(-processPid, syscall.SIGINT)
			// If checkAttempts speified check and wait to make sure process was killed.
			for i := 0; i < checkAttempts; i++ {
				ControllerDebug.log("Process check ", theProcessType, i)
				if cmps.processes[theProcessType].Signal(syscall.Signal(0)) != nil {
					break //process is dead
				} else {
					time.Sleep(2 * time.Second)
				}
			}
		}
		cmps.processes[theProcessType] = nil
		cmps.pids[theProcessType] = 0
		if err != nil {

			ControllerError.log("Killing process ", processPid, " returned an error SIGINT received error ", err)

		}
	}
	return err
}

/*
	runPrep
*/
func runPrep(commandString string, interactive bool) (*exec.Cmd, error) {
	var err error
	cmd := exec.Command("/bin/bash", "-c", commandString)
	ControllerDebug.log("Set workdir:  " + workDir)
	cmd.Dir = workDir
	if interactive {
		cmd.Stdin = os.Stdin
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	ControllerInfo.log("Running APPSODY_PREP command: " + commandString)
	err = cmd.Run()

	return cmd, err
}

/*
	StartProcess
*/
func startProcess(commandString string, theProcessType ProcessType, interactive bool) (*exec.Cmd, error) {
	var err error
	cmd := exec.Command("/bin/bash", "-c", commandString)
	ControllerDebug.log("Set workdir:  " + workDir)
	cmd.Dir = workDir
	if interactive {
		cmd.Stdin = os.Stdin
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	ControllerInfo.log("Running command:  " + commandString)
	err = cmd.Start()

	cmps.processes[theProcessType] = cmd.Process

	cmps.pids[theProcessType] = cmd.Process.Pid
	ControllerDebug.log("New process created with pid ", strconv.Itoa(cmd.Process.Pid))

	return cmd, err
}

func waitProcess(cmd *exec.Cmd, theProcessType ProcessType) error {

	err := cmd.Wait()

	return err
}

func runWatcher(fileChangeCommand string, dirs []string, killServer bool, interactive bool) error {
	errorMessage := ""
	var err error

	ControllerDebug.log("Starting watcher")
	// Start the Watcher
	// compile the regex prior to running watcher because panic leaves child processes if it occurs

	r := regexp.MustCompile(appsodyWATCHREGEX)
	w := watcher.New()
	for _, ignoredir := range appsodyWATCHIGNOREDIR {
		r1 := regexp.MustCompile("^" + ignoredir)
		w.AddFilterHook(watcher.NegativeFilterHook(r1, true))
	}
	// These filter hooks MUST be added prior do adding recursive directories
	// otherwise there is a timing window at startup and unwanted events will be proccessed.
	w.AddFilterHook(watcher.NoDirectoryFilterHook())
	w.AddFilterHook(watcher.RegexFilterHook(r, false))
	w.SetMaxEvents(1)
	for d := 0; d < len(dirs); d++ {
		// Watch each directory specified recursively for changes.
		currentDir := dirs[d]
		// Make sure the directory exists
		_, err = os.Stat(currentDir)
		if err != nil {

			errorMessage = "The directory specified for file watching does not exist: " + currentDir
			ControllerWarning.log(errorMessage, err)

		}
		if err = w.AddRecursive(currentDir); err != nil {

			errorMessage = "Failed to add directory to recursive file watching list: " + currentDir
			ControllerWarning.log(errorMessage, err)

		}

	}

	// Only files that match the regular expression during file listings
	// will be watched.  Currently we watch java, js, and go files.
	// We may add an environment variable to add to this list

	//handle the ignore dirs by using a negative filter hook

	// Start the watching process - it'll check for changes every "n" ms.

	go func() {
		for {
			select {
			case event := <-w.Event:
				ControllerDebug.log("File watch event detected for:  " + event.String())

				ControllerDebug.log("About to perform the ON_CHANGE action.")

				if fileChangeCommand != "" {
					go runCommands(fileChangeCommand, fileWatcher, killServer, false, interactive)
				}

			case err := <-w.Error:
				ControllerWarning.log("An error occured in the file watcher ", err)
			case <-w.Closed:
				ControllerDebug.log("The file watcher is now closed")
				return
			}
		}
	}()

	ControllerDebug.log("The watch interval is set to: ", appsodyWATCHINTERVAL, " seconds.")
	if err = w.Start(appsodyWATCHINTERVAL); err != nil {
		errorMessage = "Could not start the watcher "
		ControllerError.log(errorMessage+" ", err)
	}
	defer w.Close()
	// Close the watcher at function end
	//defer w.Close()

	return err
}

/*
   determine if we need to kill the server process

*/
func runCommands(commandString string, theProcessType ProcessType, killServer bool, noWatcher bool, interactive bool) {

	var cmd *exec.Cmd
	var err error
	var mutexUnlocked bool

	// Start a new watch action
	ControllerDebug.log("Running command:  "+commandString, " for process type ", processTypeToString(theProcessType))
	cmps := appsodyControllerManagedProcesses()

	// lock the mutex which protects the Process and the Pid String
	cmps.mu.Lock()

	if theProcessType == server {

		// keep going
		cmd, err = startProcess(commandString, server, interactive)
		ControllerDebug.log("Started RUN/DEBUG/TEST process")
		if err != nil {
			ControllerWarning.log("ERROR start server (APPSODY_RUN/DEBUG/TEST) received error ", err)
		}
		cmps.mu.Unlock()
		mutexUnlocked = true

		err = waitProcess(cmd, theProcessType)
		if noWatcher {
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {

					statusCode := exitErr.ExitCode()
					ControllerError.log("Wait received error with status code: " + strconv.Itoa(statusCode) + " due to error: " + err.Error())
					reapChildProcesses(5)
					os.Exit(statusCode)
					// The program has exited with an exit code != 0

				} else {
					ControllerError.log("Could not determine exit code for error: ", err)
					// run the reaper to clean up anything
					reapChildProcesses(5)
					os.Exit(1)
				}
			}
			reapChildProcesses(5)
		} else {
			if err != nil {
				ControllerInfo.log("Wait received error on APPSODY_RUN/DEBUG/TEST ", err)
			}
		}
	} else {
		ControllerDebug.log("Inside the ON_CHANGE path")
		// This is a watcher
		if killServer {
			ControllerDebug.log("APPSODY_RUN/DEBUG/TEST_ON_KILL is true, attempting to kill the corresponding process.")
			err = killProcess(server, 0)
			if err != nil {
				// do nothing we continue after kill errors
				ControllerWarning.log("The attempt to kill the process received an error ", err)
			}
		}
		ControllerDebug.log("Killing the APPSODY_RUN/DEBUG/TEST_ON_CHANGE process.")

		err = killProcess(fileWatcher, 0)
		if err != nil {
			// do nothing we continue after kill errors
			ControllerWarning.log("Killing the the APPSODY_RUN/DEBUG/TEST_ON_CHANGE process received error ", err)
		}
		go reapChildProcesses(5)

		commandToUse := commandString
		processTypeToUse := fileWatcher

		if !killServer {
			// this path is only relevant for APPSODY_<RUN/DEBUG/TEST>KILL_SERVER=FALSE
			// get the process of the current server (should not be nil ever) and send benign SIG 0 to the server proces
			if cmps.processes[server] != nil && cmps.processes[server].Signal(syscall.Signal(0)) != nil {
				// if there is no server process, an error is returned
				ControllerDebug.log("The server process with pid:", cmps.processes[server].Pid, "was not found, and APPSODY_<action>_KILL is set to false. The server will be restarted.")
				//start the server with the startCommand, not the watch action command
				commandToUse = startCommand
				processTypeToUse = server
			}
		}
		ControllerDebug.log("Starting process of type ", processTypeToString(processTypeToUse), " running command: ", commandToUse)

		cmd, err = startProcess(commandToUse, processTypeToUse, interactive)

		if err != nil {
			ControllerWarning.log("Received and error starting process of type ", processTypeToString(processTypeToUse), " running command: ", commandToUse, " error received was: ", err)

		}
		cmps.mu.Unlock()
		mutexUnlocked = true

		err = waitProcess(cmd, processTypeToUse)
		if err != nil {
			// do nothing as the kill causees and error condition
			ControllerWarning.log("Wait Received error starting process of type ", processTypeToString(processTypeToUse), " while running command: ", commandToUse, " error received was: ", err)

		}

	}

	if !mutexUnlocked {

		cmps.mu.Unlock()
	}

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
	var disableWatcher bool

	mode := flag.String("mode", "run", "This is the mode the controller runs in: run, debug or test")
	flag.BoolVar(&verbose, "verbose", false, "Turns on debug output and logging ")
	flag.BoolVar(&vmode, "v", false, "Turns on debug output and logging ")
	flag.BoolVar(&disableWatcher, "no-watcher", false, "Disable file watching regardless of environment variables.")
	flag.BoolVar(&version, "version", false, "Prints the controller version and exits")
	flag.BoolVar(&interactiveFlag, "interactive", false, "Controller runs in interactive mode")

	flag.Parse()

	klogFlags = flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	if version {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	if vmode || verbose {
		// set debug mode
		_ = klogFlags.Set("v", "4")

	}
	_ = klogFlags.Set("skip_headers", "true")

	if disableWatcher {
		ControllerInfo.log("File watching has been turned off at the request of the CLI.")
	}

	ControllerDebug.log("Running Appsody Controller version " + VERSION)

	if strings.Compare(*mode, "test") == 0 {
		testMode = true
	}

	if strings.Compare(*mode, "debug") == 0 {
		debugMode = true
	}

	workDir, errWorkDir = os.Getwd()

	if errWorkDir != nil {

		ControllerFatal.log("Could not find the working dir ", errWorkDir)
		os.Exit(1)
	}
	// Obtain the environment variables
	err = setupEnvironmentVars()
	if err != nil {
		errorMessage = "Fatal: Appsody Controller setup did not find all environment variables "
		ControllerFatal.log(errorMessage, err)
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
		ControllerWarning.log("Warning: the APPSODY_DEBUG,APPSODY_TEST or APPSODY_RUN command is unspecified")
	}
	ControllerDebug.log("APPSODY_DEBUG,APPSODY_TEST or APPSODY_RUN command : " + startCommand)
	if debugMode {
		//note this could be ""
		fileChangeCommand = appsodyDEBUGWATCHACTION
	} else if testMode {
		fileChangeCommand = appsodyTESTWATCHACTION
	} else {
		fileChangeCommand = appsodyRUNWATCHACTION
	}
	ControllerDebug.log("File change command: " + fileChangeCommand)

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		ControllerDebug.log("Inside signal handler for controller")
		ControllerDebug.log("Killing the ON_CHANGE process")
		// In practice either the fileWatcher or server process will be alive, not both
		err := killProcess(fileWatcher, 2) // we give 2 *2 seconds to kill the filewatcher/ON_CHANGE process
		if err != nil {
			ControllerError.log("Received error during signal handler killing ON_CHANGE process", err)
		}
		ControllerDebug.log("Killing the server process")
		err = killProcess(server, 2) // we give 2 *2 seconds to kill the server process
		if err != nil {
			ControllerError.log("Received error during signal handler killing the RUN/TEST/DEBUG process", err)
		}
		// 5 * .2 second waiting for reaping of child processes

		reapChildProcesses(5)
		os.Exit(0)
		ControllerDebug.log("Done processing controller signal handler.")
	}()

	if appsodyPREP != "" {
		ControllerDebug.log("Running APPSODY_PREP command: ", appsodyPREP)

		_, err = runPrep(appsodyPREP, interactiveFlag)
	}
	if err != nil {
		ControllerError.log("FATAL error APPSODY_PREP command received an error.  The controller is exiting: ", err)
		os.Exit(1)
	}

	if fileChangeCommand == "" || disableWatcher {
		ControllerDebug.log("The fileChangeCommand environment variable APPSODY_RUN/DEBUG/TEST_ON_CHANGE is unspecified or file watching was disabled by the CLI.")
		ControllerDebug.log("Running APPSODY_RUN,APPSODY_DEBUG or APPSODY_TEST sync: " + startCommand)
		runCommands(startCommand, server, false, true, interactiveFlag)

	} else {
		ControllerDebug.log("Running APPSODY_RUN,APPSODY_DEBUG or APPSODY_TEST async: " + startCommand)

		go runCommands(startCommand, server, false, false, interactiveFlag)

	}

	if fileChangeCommand != "" && !disableWatcher {

		err = runWatcher(fileChangeCommand, dirs, stopWatchServerOnChange, interactiveFlag)
	} else {

		ControllerInfo.log("The file watcher is not running because no APPSODY_RUN/TEST/DEBUG_ON_CHANGE action was specified or it has been disabled using the --no-watcher flag.")
	}
	if err != nil {
		errorMessage = "Error running the file watcher: "
		ControllerFatal.log(errorMessage, err)
		os.Exit(1)
	}

}

func reapChildProcesses(maxLimit int) {
	countLimit := 0

	for {

		var wstatus syscall.WaitStatus
		//WNOHANG means return if there are no child processes to wait for
		//This command will wait for processes that hae been reassigned
		// to pid 1 after the server or fileWatcher/ON_CHANGE process is terminated
		pid, err := syscall.Wait4(-1, &wstatus, syscall.WNOHANG, nil)
		ControllerDebug.log("Reaper pid/err is: ", pid, err)
		// If it is 0 that means no process was waiting atm, we will sleep and give a little more time
		if pid == 0 && countLimit < maxLimit && err == nil {
			ControllerDebug.log("Reaper sleeping 200 millisecond: ", pid)
			time.Sleep(200 * time.Millisecond)
			countLimit++
		}

		if syscall.EINTR == err {
			// A Signal Interupt occured and we should stop processing
			ControllerDebug.log("Signal Interrupt: ", err)
			break
		}
		//This value means no child processes left waiting.
		if syscall.ECHILD == err {
			ControllerDebug.log("No more child processes: ", err)
			break
		}

		ControllerDebug.log("Max limit count: ", countLimit)

		if countLimit >= maxLimit {
			ControllerDebug.log("Max limit reached: ", maxLimit)
			break
		}

	}

}
