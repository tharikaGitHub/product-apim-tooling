/*
*  Copyright (c) WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
*
*  WSO2 Inc. licenses this file to you under the Apache License,
*  Version 2.0 (the "License"); you may not use this file except
*  in compliance with the License.
*  You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied.  See the License for the
* specific language governing permissions and limitations
* under the License.
*/

package cmd 

import (
	"os"
	"fmt"
	"strconv"
	"os/exec"
	"bufio"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"github.com/spf13/cobra"
	"github.com/Jeffail/gabs"
	jsoniter "github.com/json-iterator/go"
	"github.com/ghodss/yaml"
	
	"github.com/wso2/product-apim-tooling/import-export-cli/box"
 
	yaml2 "gopkg.in/yaml.v2"
	
	v2 "github.com/wso2/product-apim-tooling/import-export-cli/specs/v2"
 
	"github.com/wso2/product-apim-tooling/import-export-cli/utils"
)

var flagApiNameToGet string		//name of the api to get from aws gateway
var flagStageName string		//api stage to get  
var dir string					//dir where the aws init command is executed from
var path string					//path of the OAS extracted from AWS
var tmpDir string				//temporary directory created to store the OAS extracted from aws till the project is initialized
var getRestAPIsCmdOutput string 
var err error 
var awsInitCmdForced bool

//common aws cmd flags 
var awsCmdLiteral string = "aws"
var apiGateway string = "apigateway"

//aws cmd Output type
var outputFlag string = "--output"
var outputType string = "json"

//cmd_1 
var awsCLIVersionFlag string = "--version"

//getRestAPIsCmd
var getAPI string = "get-rest-apis"

//getExportCmd
var getExport string = "get-export"
var apiIdFlag string = "--rest-api-id"
var stageNameFlag string = "--stage-name"
var exportTypeFlag string = "--export-type"
var exportType string = "oas30"	//default export type is openapi3. Use "swagger" to request for a swagger 2.
var debugFlag string			//aws cli debug flag for apictl verbose mode 
//
const awsInitCmdLiteral = "aws init"
const awsInitCmdShortDesc = "Initialize a API project from a AWS API"
const awsInitCmdLongDesc = `Downloading the OpenAPI specification of an API from the AWS API Gateway to initialize a WSO2 API project`
const awsInitCmdExamples = utils.ProjectName + ` ` + awsInitCmdLiteral  + ` -n PetStore -s Demo

` + utils.ProjectName + ` ` +  awsInitCmdLiteral + ` --name PetStore --stage Demo

` + utils.ProjectName + ` ` +  awsInitCmdLiteral + ` --name Shopping --stage Live

NOTE: Both flags --name (-n) and --stage (-s) are mandatory as both values are needed to get the openAPI from AWS API Gateway.
Make sure the API name and Stage Name are correct.
Also make sure you have AWS CLI installed and configured before executing the aws init command.
Vist https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html for more info`

func getPath() {
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	dir = pwd
}

//aws init Cmd
var AwsInitCmd = &cobra.Command{
	Use:   	awsInitCmdLiteral,
	Short: 	awsInitCmdShortDesc,
	Long:  	awsInitCmdLongDesc,
	Example: 	awsInitCmdExamples,
	Run: func(cmd *cobra.Command, args []string) {
		getPath()
		initCmdOutputDir = dir + "/" + flagApiNameToGet

		if stat, err := os.Stat(initCmdOutputDir); !os.IsNotExist(err) {
			fmt.Printf("%s already exists\n", initCmdOutputDir)
			if !stat.IsDir() {
				fmt.Printf("%s is not a directory\n", initCmdOutputDir)
				os.Exit(1)
			}
			if !awsInitCmdForced {
				fmt.Println("Run with -f or --force to overwrite directory and create project")
				os.Exit(1)
			}
			fmt.Println("Running command in forced mode")
		}
		execute()
	},
}

type Apis struct {
	Items []struct {
		Id                    string `json:"id"`
		Name                  string `json:"name"`
	} `json:"items"`
}

func getOAS() error {
	utils.Logln(utils.LogPrefixInfo + "Executing aws version command")
	//check whether aws cli is installed
	cmd_1, err := exec.Command(awsCmdLiteral, awsCLIVersionFlag).Output()
	if err != nil {
		fmt.Println("Error getting AWS CLI version. Make sure aws cli is installed and configured.")
		return err
	}
	output := string(cmd_1[:])
	utils.Logln(utils.LogPrefixInfo + "AWS CLI version :  " + output)

	if utils.VerboseModeEnabled() {
		debugFlag = "--debug"	//activating the aws cli debug flag in apictl verbose mode 
	}
	utils.Logln(utils.LogPrefixInfo + "Executing aws get-rest-apis command in debug mode")
	//
	getRestAPIsCmd := exec.Command(awsCmdLiteral, apiGateway, getAPI, outputFlag, outputType, debugFlag)
	stderr, err := getRestAPIsCmd.StderrPipe()
	if err != nil {
		fmt.Println("Error creating pipe to standard error. (get-rest-apis command)", err)
	}
	stdout, err := getRestAPIsCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating pipe to standard output (get-rest-apis command).", err)
	}

	err = getRestAPIsCmd.Start()
	if err != nil {
		fmt.Println("Error starting get-rest-apis command.", err)
	}

	if utils.VerboseModeEnabled() {
		logsScannerGetRestAPIsCmd := bufio.NewScanner(stderr)
		for logsScannerGetRestAPIsCmd.Scan() {
			fmt.Println(logsScannerGetRestAPIsCmd.Text())
		}

		if err := logsScannerGetRestAPIsCmd.Err(); err != nil {
			fmt.Println("Error reading debug logs from standard error. (get-rest-apis command)", err)
		}
	}

	outputScannerGetRestAPIsCmd := bufio.NewScanner(stdout)
	for outputScannerGetRestAPIsCmd.Scan() {
		getRestAPIsCmdOutput = getRestAPIsCmdOutput + outputScannerGetRestAPIsCmd.Text()
	}

	if err := outputScannerGetRestAPIsCmd.Err(); err != nil {
		fmt.Println("Error reading output from standard output.", err)
	}
	err = getRestAPIsCmd.Wait()
	if err != nil {
		fmt.Println("Could not complete get-rest-apis command successfully.", err)
	}

	//Unmarshal from JSON into Apis struct.
	apis := Apis{}
	err = json.Unmarshal([]byte(getRestAPIsCmdOutput), &apis)
	if err != nil {
		return err
	}
	extractedAPIs := strconv.Itoa(len(apis.Items))
	utils.Logln(utils.LogPrefixInfo + extractedAPIs + " APIs were extracted")

	var found bool
	apiName := flagApiNameToGet
	stageName := flagStageName
	path = tmpDir + "/" + apiName + ".json"
	// Searching for API ID:
	utils.Logln(utils.LogPrefixInfo + "Searching for API ID...")
	for _, item := range apis.Items {
		if item.Name == apiName {
			utils.Logln("API ID found : ", item.Id)
			api_id := item.Id 
			found = true

			utils.Logln(utils.LogPrefixInfo + "Executing aws get-export command in debug mode")
			getExportCmd:= exec.Command(awsCmdLiteral, apiGateway, getExport, apiIdFlag, api_id, stageNameFlag, stageName, exportTypeFlag, exportType, path, outputFlag, outputType, debugFlag)
			stderr, err := getExportCmd.StderrPipe()
			if err != nil {
				fmt.Println("Error creating pipe to standard error. (get-export command)", err)
			}
			stdout, err := getExportCmd.StdoutPipe()
			if err != nil {
				fmt.Println("Error creating pipe to standard output. (get-export command)", err)
			}

			err = getExportCmd.Start()
			if err != nil {
				fmt.Println("Error starting get-export command.", err)
			}

			if utils.VerboseModeEnabled() {
				logsScannerGetExportCmd := bufio.NewScanner(stderr)
				for logsScannerGetExportCmd.Scan() {
					fmt.Println(logsScannerGetExportCmd.Text())
				}
				if err := logsScannerGetExportCmd.Err(); err != nil {
					fmt.Println("Error reading debug logs from standard error. (get-export command)", err)
				}
			}
			
			if utils.VerboseModeEnabled() {
				outputScannerGetExportCmd := bufio.NewScanner(stdout)
				for outputScannerGetExportCmd.Scan() {
					fmt.Println(outputScannerGetExportCmd.Text())
				}
				if err := outputScannerGetExportCmd.Err(); err != nil {
					fmt.Println("Error reading output from standard output. (get-export command)", err)
				}
			}

			err = getExportCmd.Wait()
			if err != nil {
				fmt.Println("Could not complete get-export command successfully.", err)
			} 
			break
		}
	}
	if !found {
		fmt.Println("Unable to find an API with the name", apiName)
		os.RemoveAll(tmpDir)
		os.Exit(1)
		return err
	}
	return nil
}

// loadDefaultAWSDocFromDisk loads document.yaml stored in HOME/.wso2apictl/document.yaml
func loadDefaultAWSDocFromDisk() (*v2.Document, error) {
	docData, err := ioutil.ReadFile(utils.DefaultAWSDocFilePath)
	if err != nil {
		return nil, err
	}
	awsDoc := &v2.Document{}
	err = yaml.Unmarshal(docData, &awsDoc)
	if err != nil {
		return nil, err
	}
	return awsDoc, nil
}

func createAWSDocDirectory(docName string) error {
	awsDocDirectoryPath := initCmdOutputDir + "/Docs"
	dirPath := filepath.Join(awsDocDirectoryPath, filepath.FromSlash(docName))
	utils.Logln(utils.LogPrefixInfo + "Creating directory " + dirPath)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//write document.yaml file
func writeDocumentFile(docName string, summary string) error {
	document, err := loadDefaultAWSDocFromDisk()
	docData := &document.Data
	docData.Name = docName
	docData.Summary = summary
	docDataByte, err := yaml2.Marshal(document)
	if err != nil {
		return err
	}
	apiDocFilePath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Docs/" + docName + "/document.yaml"))
	utils.Logln(utils.LogPrefixInfo + "Writing " + apiDocFilePath)
	err = ioutil.WriteFile(apiDocFilePath, docDataByte, os.ModePerm)
	return nil
}

// write AWS API security documents based on APIs security schemes
func writeAWSSecurityDocs(oas3ByteValue []byte) error {
	securitySchemes := &v2.SecuritySchemes{}
	json.Unmarshal(oas3ByteValue, &securitySchemes)
	schemes := securitySchemes.Components.SecuritySchemes
	if securitySchemes.ResourcePolicy.Version != "" {
		docName := "Resource_Policy"
		summary := "This document contains details related to AWS resource policies"
		err = createAWSDocDirectory(docName)
		resourcePolicyDocPath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Docs/" + docName + "/" + docName))
		utils.Logln(utils.LogPrefixInfo + "Writing " + resourcePolicyDocPath)
		resourcePolicyDoc, _ := box.Get("/init/resource_policy_doc")
		err = ioutil.WriteFile(resourcePolicyDocPath, resourcePolicyDoc, os.ModePerm)
		if err != nil {
			return err
		}
		err = writeDocumentFile(docName, summary)
		if err != nil {
			return err
		}
	}
	if schemes.CognitoAuthorizer.AuthType == "cognito_user_pools" {
		docName := "Cognito_Userpool"
		summary := "This document contains details related to AWS cognito user pools"
		err = createAWSDocDirectory(docName)
		cognitoUpDocPath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Docs/" + docName + "/" + docName))
		utils.Logln(utils.LogPrefixInfo + "Writing " + cognitoUpDocPath)
		cognitoUpDoc, _ := box.Get("/init/cognito_userpool_doc")
		err = ioutil.WriteFile(cognitoUpDocPath, cognitoUpDoc, os.ModePerm)
		if err != nil {
			return err
		}
		err = writeDocumentFile(docName, summary)
		if err != nil {
			return err
		}
	}
	if schemes.APIKey.Type == "apiKey" {
		docName := "AWS_API_Keys"
		summary := "This document contains details related to AWS API keys"
		err = createAWSDocDirectory(docName)
		apiKeyDocPath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Docs/" + docName + "/" + docName))
		utils.Logln(utils.LogPrefixInfo + "Writing " + apiKeyDocPath)
		apiKeyDoc, _ := box.Get("/init/aws_apikey_doc")
		err = ioutil.WriteFile(apiKeyDocPath, apiKeyDoc, os.ModePerm)
		if err != nil {
			return err
		}
		err = writeDocumentFile(docName, summary)
		if err != nil {
			return err
		}
	}
	if schemes.Sigv4.AuthType == "awsSigv4" {
		docName := "AWS_Signature_Version4"
		summary := "This document contains details related to AWS signature version 4"
		err = createAWSDocDirectory(docName)
		awsSigv4DocPath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Docs/" + docName + "/" + docName))
		utils.Logln(utils.LogPrefixInfo + "Writing " + awsSigv4DocPath)
		awsSigv4Doc, _ := box.Get("/init/aws_sigv4_doc")
		err = ioutil.WriteFile(awsSigv4DocPath, awsSigv4Doc, os.ModePerm)
		if err != nil {
			return err
		}
		err = writeDocumentFile(docName, summary)
		if err != nil {
			return err
		}
	}
	return nil
}

func initializeProject() error {
	initCmdOutputDir = flagApiNameToGet 
	swaggerSavePath := filepath.Join(initCmdOutputDir, filepath.FromSlash("Definitions/swagger.yaml"))
	fmt.Println("Initializing a new WSO2 API Manager project in", dir)
	
	definitionFile, err := loadDefaultSpecFromDisk()
	
	// Get the API DTO specific details to process
	def := &definitionFile.Data
	if err != nil {
		return err
	}

	err = createDirectories(initCmdOutputDir)
	if err != nil {
		return err
	}

	// use swagger to auto generate
	// load swagger from tmp directory
	doc, err := loadSwagger(path)
	if err != nil {
		return err
	}

	// We use swagger2 loader. It works fine for now
	// Since we don't use 3.0 specific details its ok
	// otherwise please use v2.openAPI3 loaders
	def.IsAWSAPI = true 
	err = v2.Swagger2Populate(def, doc)
	if err != nil {
		return err
	}

	oas3ByteValue := v2.CreateEpConfigForAwsAPIs(def, path)
	v2.AddAwsTag(def)

	// convert and save swagger as yaml
	yamlSwagger, err := utils.JsonToYaml(doc.Raw())
	if err != nil {
		return err
	}

	// write to file
	err = ioutil.WriteFile(swaggerSavePath, yamlSwagger, os.ModePerm)
	if err != nil {
		return err
	}

	apiData, err := yaml2.Marshal(definitionFile)
	if err != nil {
		return err
	}

	// write to the disk
	apiJSONPath := filepath.Join(initCmdOutputDir, filepath.FromSlash("api.yaml"))
	utils.Logln(utils.LogPrefixInfo + "Writing " + apiJSONPath)
	err = ioutil.WriteFile(apiJSONPath, apiData, os.ModePerm)
	if err != nil {
		return err
	}

	apimProjReadmeFilePath := filepath.Join(initCmdOutputDir, "README.md")
	utils.Logln(utils.LogPrefixInfo + "Writing " + apimProjReadmeFilePath)
	readme, _ := box.Get("/init/README.md")
	err = ioutil.WriteFile(apimProjReadmeFilePath, readme, os.ModePerm)
	if err != nil {
		return err
	}

	// Create metaData struct using details from definition
	metaData := utils.MetaData{
		Name:    definitionFile.Data.Name,
		Version: definitionFile.Data.Version,
	}
	marshaledData, err := jsoniter.Marshal(metaData)
	if err != nil {
		return err
	}

	jsonMetaData, err := gabs.ParseJSON(marshaledData)
	metaDataContent, err := utils.JsonToYaml(jsonMetaData.Bytes())
	if err != nil {
		return err
	}
	// write api_meta.yaml file to the project directory
	apiMetaDataPath := filepath.Join(initCmdOutputDir, filepath.FromSlash(utils.MetaFileAPI))
	utils.Logln(utils.LogPrefixInfo + "Writing " + apiMetaDataPath)
	err = ioutil.WriteFile(apiMetaDataPath, metaDataContent, os.ModePerm)
	if err != nil {
		return err
	}

	err = writeAWSSecurityDocs(oas3ByteValue)
	if err != nil {
		return err
	}

	fmt.Println("Project initialized")
	fmt.Println("Open README file to learn more")
	return nil
}

//execute the aws init command 
func execute() {
	tmpDir, err = ioutil.TempDir(dir, "OAS")
		if err != nil {
			os.RemoveAll(tmpDir)
			fmt.Println("Error creating temporary directory to store OAS")
			return
		}
	utils.Logln(utils.LogPrefixInfo + "Temporary directory created")

	err = getOAS()
	if err != nil {
		os.RemoveAll(tmpDir)
		utils.HandleErrorAndExit("Error getting OAS from AWS.", err)
	}
	err = initializeProject()
	if err != nil {
		os.RemoveAll(tmpDir)
		utils.HandleErrorAndExit("Error initializing project.", err)
	}
	defer os.RemoveAll(tmpDir)
	utils.Logln(utils.LogPrefixInfo + "Temporary directory deleted")
}

func init() { 
		RootCmd.AddCommand(AwsInitCmd)		
		AwsInitCmd.Flags().StringVarP(&flagApiNameToGet, "name", "n", "", "Name of the API to get from AWS Api Gateway")
		AwsInitCmd.Flags().StringVarP(&flagStageName, "stage", "s", "", "Stage name of the API to get from AWS Api Gateway")
		AwsInitCmd.Flags().BoolVarP(&awsInitCmdForced, "force", "f", false, "Force create project")

		AwsInitCmd.MarkFlagRequired("name")
		AwsInitCmd.MarkFlagRequired("stage")
}
