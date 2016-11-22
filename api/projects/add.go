package projects

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

// Creating a new project
func add(c echo.Context) error {
	// Get DB connection
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project test add")
		return errors.New("Unable to get DB from context")
	}

	hash := c.Param("hash")

	// Validate that project identifying hash is in the DB
	if db.First(&models.Project{Hash: hash}).RecordNotFound() {
		logrus.Error("Unable to find hash in db when uploading test cases")
		return constants.UNRECOGNIZED_HASH
	}

	dir := path.Join("projects", hash)

	// Get submitted input test files
	inFile, err := c.FormFile("inFile")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("Unable to get inFile")
		return err
	}

	// Get submitted output test files
	outFile, err := c.FormFile("outFile")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("Unable to get outFile")
		return err
	}

	// Ignore errors because the path should exist
	_ = os.Mkdir(path.Join("projects", hash), 0700)

	inFileHash := md5.New()

	// Get file io.ReadCloser for input test file
	inFileReader, err := inFile.Open()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("unable to open in file")
		return err
	}
	defer inFileReader.Close()

	// Write input test file to byte buffer and md5 hasher
	var inBuf bytes.Buffer
	if _, err := io.Copy(&inBuf, io.TeeReader(inFileReader, inFileHash)); err != nil {
		logrus.WithError(err).Error("Error reading uploaded test case to md5 hasher and buffer")
		return err
	}

	// TODO figure out how to better decode the checksum
	filenameBase := string(fmt.Sprintf("%x", inFileHash.Sum(nil)))

	// Write file for holding input test file on filesystem
	inFileWriter, err := os.Create(path.Join(dir, filenameBase+".in"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"filename": path.Join(dir, filenameBase+".in"),
		}).Error("unable to open in file writer when uploading test case")
		return err
	}
	defer inFileWriter.Close()

	// Write input test file to filesystem
	if _, err = io.Copy(inFileWriter, &inBuf); err != nil {
		logrus.WithFields(logrus.Fields{
			"hash":     hash,
			"filename": path.Join(dir, filenameBase+".in"),
			"error":    err,
		}).Error("unable to write uploaded test case")
		return err
	}

	// Repeat file writing for output test files
	outFileReader, err := outFile.Open()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("unable to open out file")
		return err
	}
	defer outFileReader.Close()

	outFileWriter, err := os.Create(path.Join(dir, filenameBase+".out"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"filename": path.Join(dir, filenameBase+".in"),
		}).Error("unable to open out file writer when uploading test case")
		return err
	}
	defer outFileWriter.Close()

	if _, err = io.Copy(outFileWriter, outFileReader); err != nil {
		logrus.WithFields(logrus.Fields{
			"hash":     hash,
			"filename": path.Join(dir, filenameBase+".out"),
			"error":    err,
		}).Error("unable to write uploaded test case expected output")
		return err
	}

	// Return links to test files
	testcase := Testcase{
		InputLink:  getTestcaseLink(hash, filenameBase+".in"),
		OutputLink: getTestcaseLink(hash, filenameBase+".out"),
	}
	return c.JSON(http.StatusOK, testcase)
}
