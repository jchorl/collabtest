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

func add(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project test add")
		return errors.New("Unable to get DB from context")
	}

	hash := c.Param("hash")

	// validate that hash is in the db
	if db.First(&models.Project{Hash: hash}).RecordNotFound() {
		logrus.Error("Unable to find hash in db when uploading test cases")
		return constants.UNRECOGNIZED_HASH
	}

	dir := path.Join("projects", hash)

	inFile, err := c.FormFile("inFile")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("Unable to get inFile")
		return err
	}

	outFile, err := c.FormFile("outFile")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("Unable to get outFile")
		return err
	}

	// ignore errors because the path should exist
	_ = os.Mkdir(path.Join("projects", hash), 0700)

	inFileHash := md5.New()

	inFileReader, err := inFile.Open()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err,
			"context": c,
		}).Error("unable to open in file")
		return err
	}
	defer inFileReader.Close()

	// TODO accept emtpy files, currently the md5 bombs
	inFileReaderTee := io.TeeReader(inFileReader, inFileHash)
	var inBts []byte
	if _, err := inFileReaderTee.Read(inBts); err != nil {
		logrus.WithError(err).Error("Error reading uploaded test case to md5 hasher and byte slice")
		return err
	}

	// TODO figure out how to better decode the checksum
	filenameBase := string(fmt.Sprintf("%x", inFileHash.Sum(nil)))

	inFileWriter, err := os.Create(path.Join(dir, filenameBase+".in"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"filename": path.Join(dir, filenameBase+".in"),
		}).Error("unable to open in file writer when uploading test case")
		return err
	}
	defer inFileWriter.Close()

	if _, err = io.Copy(inFileWriter, bytes.NewReader(inBts)); err != nil {
		logrus.WithFields(logrus.Fields{
			"hash":     hash,
			"filename": path.Join(dir, filenameBase+".in"),
			"error":    err,
		}).Error("unable to write uploaded test case")
		return err
	}

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

	return c.NoContent(http.StatusAccepted)
}
