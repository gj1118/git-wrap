package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	copyDir "github.com/otiai10/copy"
)

// constants
const CONFIG_FILE_NAME string = "l1onResources.json"

// types
type UserConfig struct {
	RepoURL             string `json:"repo_url"`
	DestinationPath     string `json:"destination_path"`
	TempDirectory       string `json:"temp_directory"`
	DeleteTempDirectory bool   `json:"delete_temp_dir_after_done"`
	ProjectName         string `json:"project_name"`
	PurgeDestination    bool   `json:"purge_destination_before_copy"`
}

// helpers

/*
*
Check if the file exists.
We try to validate that the file exists by opening the file.
if we were able to open the file succcessfully, that would mean that the file exists.
otherwise the file does not exist.

Expects
1. path - Path to the file

Returns
1. boolean - If the file was successfully opened for a read operation or not.
*
*/
func checkIfFileExists(path string) bool {
	file, error := os.Open(path)
	if error != nil {
		return false
	}

	file.Close()
	return true
}

/*
*
Delete a directory and all its contents

Expects
1. path - Path to the directory

Returns
1. boolean - If the directory was successfully deleted or not.
*
*/
func deleteDirectory(path string) bool {
	Log("Deleting the directory: "+path, false)
	err := os.RemoveAll(path)
	if err != nil {
		Warning("Error while deleting the directory: " + path)
		return false
	}
	return true
}

/*
*
Clone the repository in the temp directory

Expects
1. repoURL - URL of the repository
2. directory - Directory where the repository will be cloned

Returns
1. boolean - If the repository was successfully cloned or not.
*
*/
func cloneRepository(repoName string, directory string) bool {
	Log("Cloning the repository: "+repoName, false)
	currDirectory, _ := os.Getwd()
	cmd := exec.Command("git", "clone", repoName, directory)
	err := cmd.Run()

	if err != nil {
		Log("Will try doing a git pull instead of git clone", true)
		os.Chdir(directory)
		cmd := exec.Command("git", "pull")
		err := cmd.Run()
		os.Chdir(currDirectory)
		if err != nil {
			Warning("Error while cloning the repository: " + repoName)
			return false
		}
	}
	return true
}

/*
*
Check if a directory exists.

Expects
1. path - Path to the directory

Returns
1. boolean - If the directory exists or not.
*
*/
func checkIfDirectoryExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

/*
*
Check if the directory exists. If the directory does not exist, we will create the directory. Otherwise we will use the directory

Expects
1. path - Path to the directory

Returns
1. boolean - If the directory was successfully created or not.

*
*/
func createDirectoryIfNotExists(path string) bool {
	Log("Validating if we need to create the new directory : "+path, false)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			Warning("Error while creating the directory: " + path)
			return false
		}
	}
	return true
}

/*
* Our billion dollar but inbuilt Log function

	Expects
	1. message - the message to be printed

	Returns
	1. void

*
*/
func Log(message string, tab bool) {
	if tab {
		fmt.Println("> " + message)
	} else {
		fmt.Println(message)
	}
}

/*
* Our billion dollar but inbuilt Info function

	Expects
	1. message - the message to be printed in color

	Returns
	1. void

*
*/
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

/*
* Our billion dollar but inbuilt Warning function

	Expects
	1. message - the message to be printed in color

	Returns
	1. void

*
*/
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func readConfigFile() UserConfig {
	Log("Will validate if the config file exists or not.", false)
	var userConfig UserConfig

	if checkIfFileExists(CONFIG_FILE_NAME) {
		Log("Config file exists in the current folder.", true)
		configFile, err := os.Open(CONFIG_FILE_NAME)
		if err != nil {
			// this should not happen as we have already validated that the file exists.
			// not sure what to do here.
			// but for now, we will again let the user know
			Warning("Config file might not exist as we are not able to read it.")
		}

		defer configFile.Close()
		byteArray, _ := ioutil.ReadAll(configFile)
		json.Unmarshal(byteArray, &userConfig)

	} else {
		Warning("Config file does NOT exist in the current folder.")
	}
	return userConfig
}

func main() {
	Log("Starting....", false)
	userConfig := readConfigFile()
	if (userConfig != UserConfig{}) {
		//config file was successfully read and the struct was populated.
		Log("Config file was successfully read and the struct was populated.", true)
		Log("Repo URL: "+userConfig.RepoURL, true)
		Log("Destination Path: "+userConfig.DestinationPath, true)
		Log("Temp Directory: "+userConfig.TempDirectory, true)
		Log("Delete Temp Directory: "+fmt.Sprint(userConfig.DeleteTempDirectory), true)
		Log("Project Name: "+userConfig.ProjectName, true)
		Log("Purge Destination: "+fmt.Sprint(userConfig.PurgeDestination), true)

		// lets start reading the temporary directory
		// we use this temporary directory to clone the repository
		tempDirectoryaVal := createDirectoryIfNotExists(userConfig.TempDirectory)
		if tempDirectoryaVal {
			Log("Temp Directory was created successfully.", true)
			// git clone the repository in the temp directory
			directoryClonedSuccessFully := cloneRepository(userConfig.RepoURL, userConfig.TempDirectory)
			if !directoryClonedSuccessFully {
				Warning("Error while cloning the repository. Please check the logs.")
				os.Exit(1)
			}
			Log("Cloned the repository successfully: "+fmt.Sprint(directoryClonedSuccessFully), true)
			Log("Prep the copy process", false)
			sourceDir := path.Join(userConfig.TempDirectory, userConfig.ProjectName)
			Log("Generated  source directory: "+sourceDir, true)
			if checkIfDirectoryExists(userConfig.DestinationPath) {
				Log("Destination directory exists.", true)
			} else {
				Log("Destination directory does NOT exist. Will attempt to create the destination directory", true)
				createDirectoryIfNotExists(userConfig.DestinationPath)
			}
			// check if we need to purge the destination directory first
			if userConfig.PurgeDestination {
				Log("Purge the destination directory", true)
				deleteError := deleteDirectory(userConfig.DestinationPath)
				if deleteError {
					Log("Purging the destination directory has happened succesfully", true)
					Log("Create the destination directory : "+userConfig.DestinationPath, true)
					destinationPathSuccess := createDirectoryIfNotExists(userConfig.DestinationPath)
					if destinationPathSuccess {
						Log("Destination directory has been created", true)
					} else {
						Log("Destination directory could not be created : "+userConfig.DestinationPath, true)
						os.Exit(1)
					}
				} else {
					Warning("Error while purging the destination directory")
					os.Exit(1)
				}
			}
			copyError := copyDir.Copy(sourceDir, userConfig.DestinationPath)
			if copyError != nil {
				Log("Error while copying the files: "+copyError.Error(), true)
				os.Exit(1)
			}
			Log("Files were copied successfully.", true)
			if userConfig.DeleteTempDirectory {
				Log("Will now delete the directory: "+userConfig.DestinationPath, false)
				// lets delete the directory now
				deleteDirectory(userConfig.TempDirectory)
			} else {
				Log("Directory cleanup will not happen", false)
			}
			Log("All done. Exiting now.", false)
			os.Exit(0)
		} else {
			Warning("Temp Directory was NOT created successfully, aborting!")
			os.Exit(1)
		}
	} else {
		//config file was not read successfully.
		//exit, preserve some dignity.
		os.Exit(1)
	}
}
