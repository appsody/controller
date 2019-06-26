package test

// Copyright © 2019 The Appsody Authors
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
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// RunBashCmdExec runs the bash command with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the bash command
// workingDir will be the directory the command runs in
func RunBashCmdExec(args []string, workingDir string) (string, error) {
	execDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer func() {
		// replace the original working directory when this funciton completes
		err := os.Chdir(execDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// set the working directory
	if err := os.Chdir(workingDir); err != nil {
		return "", err
	}
	log.Println("executing:  " + args[0])
	commandString := args[0]
	execCmd := exec.Command("/bin/sh", "-c", commandString)

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer func() {
		outReader.Close()
		outWriter.Close()
	}()
	execCmd.Stdout = outWriter
	execCmd.Stderr = outWriter
	outScanner := bufio.NewScanner(outReader)
	var outBuffer bytes.Buffer
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
			fmt.Println(string(out))
		}
	}()

	err = execCmd.Start()
	if err != nil {
		return "", err
	}

	err = execCmd.Wait()
	return outBuffer.String(), err
}

// RunBashCmdExecAndKill runs the bash command with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the bash command
// workingDir will be the directory the command runs in
func RunBashCmdExecAndKill(args []string, workingDir string, projectDir string) (string, error) {
	execDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer func() {
		// replace the original working directory when this funciton completes
		err := os.Chdir(execDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// set the working directory
	if err := os.Chdir(workingDir); err != nil {
		return "", err
	}
	log.Println("executing:  " + args[0])
	commandString := args[0]
	execCmd := exec.Command("/bin/sh", "-c", commandString)

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer func() {
		outReader.Close()
		outWriter.Close()
	}()
	execCmd.Stdout = outWriter
	execCmd.Stderr = outWriter
	outScanner := bufio.NewScanner(outReader)
	var outBuffer bytes.Buffer
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
			fmt.Println(string(out))
		}
	}()

	err = execCmd.Start()
	if err != nil {
		return "", err
	}
	time.Sleep(30 * time.Second)
	log.Printf("About to kill: %v\n", execCmd.Process.Pid)
	err = syscall.Kill(-execCmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		log.Printf("err for kill 1 %v\n", err)
	}
	err = execCmd.Process.Kill()
	if err != nil {
		log.Printf("err for kill 2 %v\n", err)
	}
	err = execCmd.Wait()
	if err != nil {
		log.Printf("err for wait %v\n", err)
	}
	execCmdRMDIR := exec.Command("/bin/sh", "-c", "rm -r +"+projectDir)
	err = execCmdRMDIR.Start()
	if err != nil {
		log.Printf("err for kill start %v\n", err)
	}
	err = execCmdRMDIR.Wait()
	log.Printf("err for wait for rmdir %v\n", err)
	return outBuffer.String(), err
}

// RunBashCmdExecAndKillAndTouch runs the bash command with the given args in a new process
// The stdout and stderr are captured, printed, and returned
// args will be passed to the bash command
// workingDir will be the directory the command runs in
func RunBashCmdExecAndKillAndTouch(args []string, workingDir string, kill bool, projectDir string) (string, error) {
	execDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer func() {
		// replace the original working directory when this funciton completes
		err := os.Chdir(execDir)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// set the working directory
	if err := os.Chdir(workingDir); err != nil {
		return "", err
	}
	log.Println("executing:  " + args[0])
	commandString := args[0]
	execCmd := exec.Command("/bin/sh", "-c", commandString)

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer func() {
		outReader.Close()
		outWriter.Close()
	}()
	execCmd.Stdout = outWriter
	execCmd.Stderr = outWriter
	outScanner := bufio.NewScanner(outReader)
	var outBuffer bytes.Buffer
	go func() {
		for outScanner.Scan() {
			out := outScanner.Bytes()
			outBuffer.Write(out)
			outBuffer.WriteByte('\n')
			fmt.Println(string(out))
		}
	}()

	err = execCmd.Start()
	if err != nil {
		return "", err
	}
	time.Sleep(10 * time.Second)
	execCmd2 := exec.Command("/bin/sh", "-c", "touch "+projectDir+"/new.go;touch "+projectDir+"/new.java;touch "+projectDir+"/new.bad")
	err = execCmd2.Run()
	if err != nil {
		log.Printf("err for run touch %v\n", err)
	}

	time.Sleep(60 * time.Second)
	if kill {
		log.Printf("About to kill: %v\n", execCmd.Process.Pid)
		err = syscall.Kill(-execCmd.Process.Pid, syscall.SIGKILL)
		if err != nil {
			log.Printf("Process Kill finished with : %v\n", err)
		}
		err = execCmd.Process.Kill()
		if err != nil {
			log.Printf("Process Kill finished with : %v\n", err)
		}
	}
	err = execCmd.Wait()
	if err != nil {
		log.Printf("err for subprocess %v\n", err)
	}

	return outBuffer.String(), err
}
