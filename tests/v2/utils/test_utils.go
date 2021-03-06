package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	schema "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"sigs.k8s.io/yaml"
)

const (
	defaultTempDir = "./tmp/"
	logFileName    = "test.log"
	// logToFileOnly - If set to false the log output will also be output to the console
	logToFileOnly = true // If set to false the log output will also be output to the console
)

// tmpDir temporary directory in use
var tmpDir string

var (
	testLogger *log.Logger
)

// init creates:
//    - the temporary directory used by the test to store logs and generated devfiles.
//    - the log file
func init() {
	tmpDir = defaultTempDir
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		os.RemoveAll(tmpDir)
	}
	if err := os.Mkdir(tmpDir, 0755); err != nil {
		fmt.Printf("Failed to create temp directory, will use current directory : %v ", err)
		tmpDir = "./"
	}
	f, err := os.OpenFile(filepath.Join(tmpDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error creating Log file : %v", err)
	} else {
		if logToFileOnly {
			testLogger = log.New(f, "", log.LstdFlags|log.Lmicroseconds)
		} else {
			writer := io.MultiWriter(f, os.Stdout)
			testLogger = log.New(writer, "", log.LstdFlags|log.Lmicroseconds)
		}
		testLogger.Println("Test Starting:")
	}

}

// CreateTempDir creates a specified sub directory under the temp directory if it does not exist.
// Returns the name of the created directory.
func CreateTempDir(subdir string) string {
	tempDir := tmpDir + subdir + "/"
	var err error
	if _, err = os.Stat(tempDir); os.IsNotExist(err) {
		err = os.Mkdir(tempDir, 0755)
	}
	if err != nil {
		// if cannot create subdirectory just use the base tmp directory
		LogErrorMessage(fmt.Sprintf("Failed to create temp directory %s will use %s : %v", tempDir, tmpDir, err))
		tempDir = tmpDir
	}
	return tempDir
}

// GetDevFileName returns a qualified name of a devfile for use in a test.
// The devfile will be in a temporary directory and is named using the calling function's name.
func GetDevFileName() string {
	pc, fn, _, ok := runtime.Caller(1)
	if !ok {
		return tmpDir + "DefaultDevfile"
	}

	testFile := filepath.Base(fn)
	testFileExtension := filepath.Ext(testFile)
	subdir := testFile[0 : len(testFile)-len(testFileExtension)]
	destDir := CreateTempDir(subdir)
	callerName := runtime.FuncForPC(pc).Name()
	pos1 := strings.LastIndex(callerName, "/parserTest.") + len("/parserTest.")
	devfileName := destDir + callerName[pos1:len(callerName)] + ".yaml"

	LogInfoMessage(fmt.Sprintf("GetDevFileName : %s", devfileName))

	return devfileName
}

// AddSuffixToFileName adds a specified suffix to the name of a specified file.
// For example if the file is devfile.yaml and the suffix is 1, the result is devfile1.yaml
func AddSuffixToFileName(fileName string, suffix string) string {
	pos1 := strings.LastIndex(fileName, ".yaml")
	newFileName := fileName[0:pos1] + suffix + ".yaml"
	LogInfoMessage(fmt.Sprintf("Add suffix %s to fileName %s : %s", suffix, fileName, newFileName))
	return newFileName
}

// LogMessage logs the specified message and returns the message logged
func LogMessage(message string) string {
	if testLogger != nil {
		testLogger.Println(message)
	} else {
		fmt.Printf("Logger not available: %s", message)
	}
	return message
}

var errorPrefix = "..... ERROR : "
var infoPrefix = "INFO :"

// LogErrorMessage logs the specified message as an error message and returns the message logged
func LogErrorMessage(message string) string {
	var errMessage []string
	errMessage = append(errMessage, errorPrefix, message)
	return LogMessage(fmt.Sprint(errMessage))
}

// LogInfoMessage logs the specified message as an info message and returns the message logged
func LogInfoMessage(message string) string {
	var infoMessage []string
	infoMessage = append(infoMessage, infoPrefix, message)
	return LogMessage(fmt.Sprint(infoMessage))
}

// TestDevfile is a structure used to track a test devfile and its contents
type TestDevfile struct {
	SchemaDevFile schema.Devfile
	FileName      string
	ParserData    devfileData.DevfileData
	SchemaParsed  bool
	GroupDefaults map[schema.CommandGroupKind]bool
	UsedPorts     map[int]bool
}

var StringCount int = 0

var RndSeed int64 = time.Now().UnixNano()

// GetRandomUniqueString returns a unique random string which is n characters long plus an integer to ensure uniqueness
// If lower is set to true a lower case string is returned.
func GetRandomUniqueString(n int, lower bool) string {
	StringCount++
	return fmt.Sprintf("%s%04d", GetRandomString(n, lower), StringCount)
}

// Creates a unique seed for the randon generation.
func setRandSeed() {
	RndSeed++
	rand.Seed(RndSeed)
}

const schemaBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GetRandomString returns a random string which is n characters long.
// If lower is set to true a lower case string is returned.
func GetRandomString(n int, lower bool) string {
	setRandSeed()
	b := make([]byte, n)
	for i := range b {
		b[i] = schemaBytes[rand.Intn(len(schemaBytes)-1)]
	}
	randomString := string(b)
	if lower {
		randomString = strings.ToLower(randomString)
	}
	return randomString
}

var GroupKinds = [...]schema.CommandGroupKind{schema.BuildCommandGroupKind, schema.RunCommandGroupKind, schema.TestCommandGroupKind, schema.DebugCommandGroupKind}

// GetRandomGroupKind return random group kind. One of "build", "run", "test" or "debug"
func GetRandomGroupKind() schema.CommandGroupKind {
	return GroupKinds[GetRandomNumber(len(GroupKinds))-1]
}

// GetBinaryDecision randomly returns true or false
func GetBinaryDecision() bool {
	return GetRandomDecision(1, 1)
}

// GetRandomDecision randomly returns true or false, but weighted to one or the other.
// For example if success is set to 2 and failure to 1, true is twice as likely to be returned.
func GetRandomDecision(success int, failure int) bool {
	setRandSeed()
	return rand.Intn(success+failure) > failure-1
}

// GetRandomNumber randomly returns an integer between 1 and the number specified.
func GetRandomNumber(max int) int {
	setRandSeed()
	return rand.Intn(max) + 1
}

// GetDevfile returns a structure used to represent a specific devfile in a test
func GetDevfile(fileName string) (TestDevfile, error) {

	var err error
	testDevfile := TestDevfile{}
	testDevfile.SchemaDevFile = schema.Devfile{}
	testDevfile.FileName = fileName
	testDevfile.SchemaDevFile.SchemaVersion = "2.0.0"
	testDevfile.SchemaParsed = false
	testDevfile.GroupDefaults = make(map[schema.CommandGroupKind]bool)
	for _, kind := range GroupKinds {
		testDevfile.GroupDefaults[kind] = false
	}
	testDevfile.ParserData, err = devfileData.NewDevfileData(testDevfile.SchemaDevFile.SchemaVersion)
	if err != nil {
		return testDevfile, err
	}
	testDevfile.ParserData.SetSchemaVersion(testDevfile.SchemaDevFile.SchemaVersion)
	testDevfile.UsedPorts = make(map[int]bool)
	return testDevfile, err
}

// WriteDevfile create a devifle on disk for use in a test.
// If useParser is true the parser library is used to generate the file, otherwise "sigs.k8s.io/yaml" is used.
func (devfile *TestDevfile) WriteDevfile(useParser bool) error {
	var err error

	fileName := devfile.FileName
	if !strings.HasSuffix(fileName, ".yaml") {
		fileName += ".yaml"
	}

	if useParser {
		LogInfoMessage(fmt.Sprintf("Use Parser to write devfile %s", fileName))

		ctx := devfileCtx.NewDevfileCtx(fileName)

		err = ctx.SetAbsPath()
		if err != nil {
			LogErrorMessage(fmt.Sprintf("Setting devfile path : %v", err))
		} else {
			devObj := parser.DevfileObj{
				Ctx:  ctx,
				Data: devfile.ParserData,
			}
			err = devObj.WriteYamlDevfile()
			if err != nil {
				LogErrorMessage(fmt.Sprintf("Writing devfile : %v", err))
			} else {
				devfile.SchemaParsed = false
			}
		}

	} else {
		LogInfoMessage(fmt.Sprintf("Marshall and write devfile %s", devfile.FileName))
		c, marshallErr := yaml.Marshal(&(devfile.SchemaDevFile))

		if marshallErr != nil {
			err = errors.New(LogErrorMessage(fmt.Sprintf("Marshall devfile %s : %v", devfile.FileName, marshallErr)))
		} else {
			err = ioutil.WriteFile(fileName, c, 0644)
			if err != nil {
				LogErrorMessage(fmt.Sprintf("Write devfile %s : %v", devfile.FileName, err))
			} else {
				devfile.SchemaParsed = false
			}
		}
	}
	return err
}

// parseSchema uses the parser to parse a devfile on disk
func (devfile *TestDevfile) parseSchema() error {

	var err error
	if !devfile.SchemaParsed {
		err = devfile.WriteDevfile(true)
		if err != nil {
			LogErrorMessage(fmt.Sprintf("From WriteDevfile %v : ", err))
		} else {
			LogInfoMessage(fmt.Sprintf("Parse and Validate %s : ", devfile.FileName))
			parsedSchemaObj, parse_err := devfilepkg.ParseAndValidate(devfile.FileName)
			if parse_err != nil {
				err = parse_err
				LogErrorMessage(fmt.Sprintf("From ParseAndValidate %v : ", err))
			}
			devfile.SchemaParsed = true
			devfile.ParserData = parsedSchemaObj.Data
		}
	}
	return err
}

// Verify verifies the contents of the specified devfile with the expected content
func (devfile *TestDevfile) Verify() error {

	LogInfoMessage(fmt.Sprintf("Verify %s : ", devfile.FileName))

	var errorString []string

	err := devfile.parseSchema()

	if err != nil {
		errorString = append(errorString, LogErrorMessage(fmt.Sprintf("parsing schema %s : %v", devfile.FileName, err)))
	} else {
		LogInfoMessage(fmt.Sprintf("Get commands %s : ", devfile.FileName))
		commands, _ := devfile.ParserData.GetCommands(common.DevfileOptions{})
		if commands != nil && len(commands) > 0 {
			err = devfile.VerifyCommands(commands)
			if err != nil {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Verfify Commands %s : %v", devfile.FileName, err)))
			}
		} else {
			LogInfoMessage(fmt.Sprintf("No command found in %s : ", devfile.FileName))
		}

		LogInfoMessage(fmt.Sprintf("Get components %s : ", devfile.FileName))
		components, _ := devfile.ParserData.GetComponents(common.DevfileOptions{})
		if components != nil && len(components) > 0 {
			err = devfile.VerifyComponents(components)
			if err != nil {
				errorString = append(errorString, LogErrorMessage(fmt.Sprintf("Verfify Commands %s : %v", devfile.FileName, err)))
			}
		} else {
			LogInfoMessage(fmt.Sprintf("No components found in %s : ", devfile.FileName))
		}
	}
	var returnError error
	if len(errorString) > 0 {
		returnError = errors.New(fmt.Sprint(errorString))
	}
	return returnError

}

// EditCommands modifies random attributes for each of the commands in the devfile.
func (devfile *TestDevfile) EditCommands() error {

	LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	err := devfile.parseSchema()
	if err != nil {
		LogErrorMessage(fmt.Sprintf("From parser : %v", err))
	} else {
		LogInfoMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
		commands, _ := devfile.ParserData.GetCommands(common.DevfileOptions{})
		for _, command := range commands {
			err = devfile.UpdateCommand(command.Id)
			if err != nil {
				LogErrorMessage(fmt.Sprintf("Updating command : %v", err))
			}
		}
	}
	return err
}

// EditComponents modifies random attributes for each of the components in the devfile.
func (devfile *TestDevfile) EditComponents() error {

	LogInfoMessage(fmt.Sprintf("Edit %s : ", devfile.FileName))

	err := devfile.parseSchema()
	if err != nil {
		LogErrorMessage(fmt.Sprintf("From parser : %v", err))
	} else {
		LogInfoMessage(fmt.Sprintf(" -> Get commands %s : ", devfile.FileName))
		components, _ := devfile.ParserData.GetComponents(common.DevfileOptions{})
		for _, component := range components {
			err = devfile.UpdateComponent(component.Name)
			if err != nil {
				LogErrorMessage(fmt.Sprintf("Updating component : %v", err))
			}
		}
	}
	return err
}
