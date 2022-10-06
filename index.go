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
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

// constants
const CONFIG_FILE_NAME string = "l1onResources.json"
const VERSION string = "4.0.0"

// types

type Projects struct {
	Projects []Project `json:"projects"`
}
type Project struct {
	RepoURL             string `json:"repo_url"`
	DestinationPath     string `json:"destination_path"`
	TempDirectory       string `json:"temp_directory"`
	DeleteTempDirectory bool   `json:"delete_temp_dir_after_done"`
	ProjectName         string `json:"project_name"`
	PurgeDestination    bool   `json:"purge_destination_before_copy"`
}

// helpers

func generateWelcomeHeader() {
	pterm.DefaultCenter.WithCenterEachLineSeparately().Println("Welcome to GIT-WRAP!\n" + VERSION)
	// Generate BigLetters
	s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("GIT-WRAP")).Srender()
	pterm.DefaultCenter.Println(s) // Print BigLetters with the default CenterPrinter

	pterm.DefaultCenter.WithCenterEachLineSeparately().Println("👋 Please make sure that the config file \nl1onResources.json \nis in the same directory as this executable.")
}

func generateSectionHeader(sectionHeader string) {
	pterm.DefaultSection.Println(sectionHeader)
}

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
	Info("Deleting the directory: " + path)
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
	Info("Cloning the repository: " + repoName)
	currDirectory, _ := os.Getwd()
	cmd := exec.Command("git", "clone", repoName, directory)
	err := cmd.Run()

	if err != nil {
		Info("Will try doing a git pull instead of git clone")
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
	Info("Validating if we need to create the new directory : " + path)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			Warning("Error while creating the directory: " + path)
			Warning("Error: " + err.Error())
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
	//fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
	pterm.Info.Println(fmt.Sprintf(format, args...))
}

func Debug(format string, args ...interface{}) {
	//fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
	pterm.Debug.Println(fmt.Sprintf(format, args...))
}

func Success(format string, args ...interface{}) {
	//fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
	pterm.Success.Println(fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	pterm.Error.Println((fmt.Sprintf(format, args...)))
}

func Fatal(format string, args ...interface{}) {
	pterm.Fatal.Println((fmt.Sprintf(format, args...)))
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

func readConfigFile() Projects {
	generateSectionHeader("Validate the config file")
	var projects Projects

	if checkIfFileExists(CONFIG_FILE_NAME) {
		Success("Config file exists in the current folder.")
		configFile, err := os.Open(CONFIG_FILE_NAME)
		if err != nil {
			// this should not happen as we have already validated that the file exists.
			// not sure what to do here.
			// but for now, we will again let the user know
			Error("Config file might not exist as we are not able to read it.")
		}

		defer configFile.Close()
		byteArray, _ := ioutil.ReadAll(configFile)
		json.Unmarshal(byteArray, &projects)

	} else {
		Error("Config file does NOT exist in the current folder.")
	}
	return projects
}

func main() {
	pterm.EnableDebugMessages() // Enable debug messages
	generateWelcomeHeader()
	projects := readConfigFile()
	if len(projects.Projects) > 0 {
		Log("Config file was successfully read and the struct was populated.", true)
		Log("There are "+fmt.Sprint(len(projects.Projects), " projects"), true)
		for i := 0; i < len(projects.Projects); i++ {
			project := projects.Projects[i]
			//config file was successfully read and the struct was populated.
			generateSectionHeader("Project Name: " + project.ProjectName)

			Log("Repo URL: "+project.RepoURL, true)
			Log("Destination Path: "+project.DestinationPath, true)
			Log("Temp Directory: "+project.TempDirectory, true)
			Log("Delete Temp Directory: "+fmt.Sprint(project.DeleteTempDirectory), true)
			Log("Project Name: "+project.ProjectName, true)
			Log("Purge Destination: "+fmt.Sprint(project.PurgeDestination), true)

			// lets start reading the temporary directory
			// we use this temporary directory to clone the repository
			tempDirectoryaVal := createDirectoryIfNotExists(project.TempDirectory)
			if tempDirectoryaVal {
				Success("Temp Directory was created successfully.")
				// git clone the repository in the temp directory
				directoryClonedSuccessFully := cloneRepository(project.RepoURL, project.TempDirectory)
				if !directoryClonedSuccessFully {
					Warning("Error while cloning the repository. Please check the logs.")
					os.Exit(1)
				}
				Success("Cloned the repository successfully: " + fmt.Sprint(directoryClonedSuccessFully))
				Info("Prep the copy process")
				sourceDir := path.Join(project.TempDirectory, project.ProjectName)
				Success("Generated  source directory: " + sourceDir)
				if checkIfDirectoryExists(project.DestinationPath) {
					Info("Destination directory exists.")
				} else {
					Log("Destination directory does NOT exist. Will attempt to create the destination directory", true)
					createDirectoryIfNotExists(project.DestinationPath)
				}
				// check if we need to purge the destination directory first
				if project.PurgeDestination {
					Log("Purge the destination directory", true)
					deleteError := deleteDirectory(project.DestinationPath)
					if deleteError {
						Log("Purging the destination directory has happened succesfully", true)
						Log("Create the destination directory : "+project.DestinationPath, true)
						destinationPathSuccess := createDirectoryIfNotExists(project.DestinationPath)
						if destinationPathSuccess {
							Log("Destination directory has been created", true)
						} else {
							Log("Destination directory could not be created : "+project.DestinationPath, true)
							os.Exit(1)
						}
					} else {
						Warning("Error while purging the destination directory")
						os.Exit(1)
					}
				}
				copyError := copyDir.Copy(sourceDir, project.DestinationPath)
				if copyError != nil {
					Log("Error while copying the files: "+copyError.Error(), true)
					os.Exit(1)
				}
				Success("Files were copied successfully.")
				if project.DeleteTempDirectory {
					Info("Will now delete the directory: " + project.DestinationPath)
					// lets delete the directory now
					deleteDirectory(project.TempDirectory)
				} else {
					Info("Directory cleanup will not happen")
				}
				Success("Finished processing the project: " + project.ProjectName)
			} else {
				Warning("Temp Directory was NOT created successfully, aborting!")
				os.Exit(1)
			}
		}
		generateSectionHeader("All done. Exiting now.")
		os.Exit(0)
	} else {
		//config file was not read successfully.
		//exit, preserve some dignity.
		os.Exit(1)
	}
}
