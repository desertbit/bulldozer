/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func IsDir(path string) (bool, error) {
	// Check if the dir exist.
	d, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// Check if the path is a directory.
	return d.IsDir(), nil
}

// Exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MkDirIfNotExists creates a directory if it doesn't exists
func MkDirIfNotExists(path string) error {
	e, err := Exists(path)
	if err != nil {
		return err
	}

	if !e {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

func CopyFileIfNotExists(source string, dest string) (err error) {
	e, err := Exists(dest)
	if err != nil {
		return err
	}

	if !e {
		return CopyFile(source, dest)
	}

	return nil
}

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}

	return
}

// Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Destination already exists")
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}
