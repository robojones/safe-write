// Package safe provides methods to manage configuration files where it is important that they are either completely
// written to the disk or not at all – even when the process is unexpectedly interrupted or there are concurrent writes.
package safe

import (
	"io/ioutil"
	"os"
	"time"
)

const AltNamePostfix = ".1"
const TimestampFormat = ".2006-01-02T15-04-05.000000"

// ReadFile removes the file with the name or $(name).1
// NotExist errors are ignored.
func RemoveFile(name string) error {
	alt := name + AltNamePostfix
	if err := remove(name); err != nil {
		return err
	}
	return remove(alt)
}

// remove a file but ignore NotExist errors
func remove(name string) error {
	err := os.Remove(name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ReadFile reads the contents of the file with the name or $(name).1
// It automatically retries three times if the files don't exist in case they are replaced concurrently.
func ReadFile(name string) ([]byte, error) {
	alt := name + AltNamePostfix
	var (
		data []byte
		err error
	)

	for i := 0; i < 3; i++ {
		data, err = ioutil.ReadFile(name)
		if !os.IsNotExist(err) {
			return data, err
		}
		data, err = ioutil.ReadFile(alt)
		if !os.IsNotExist(err) {
			return data, err
		}
	}

	return data, err
}

// WriteFilePerm writes data to a file with the provided name.
// This method also creates a temporary file which is deleted immediately after the write is complete.
// It also creates a file $(name).1 which is used to make the write concurrency and interrupt safe.
func WriteFile(name string, data []byte) error {
	return WriteFilePerm(name, 0600, data)
}

// WriteFilePerm writes data to a file with the provided name and permissions.
// This method also creates a temporary file which is deleted immediately after the write is complete.
// It also creates a file $(name).1 which is used to make the write concurrency and interrupt safe.
func WriteFilePerm(name string, perm os.FileMode, data []byte) error {
	t := time.Now()

	tmp := name + t.Format(TimestampFormat)
	alt := name + AltNamePostfix

	err := write(tmp, perm, data)
	defer os.Remove(tmp)
	if err != nil {
		return err
	}
	return safelink(tmp, alt, name)
}

// safelink creates hard links from the tmpname to the altname and from the altname to the name.
// In case a previous process was interrupted, the altname is first linked to the name.
// This complicated procedure makes sure that even if a process is interrupted before creating the link to the name,
// the the contents of the file are never lost.
func safelink(tmpname string, altname string, name string) error {
	// Attempt final link in case a previous process was interrupted before the final link.
	if err := link(altname, name); err != nil {
		return err
	}
	// Do alt link from tmp file.
	if err := link(tmpname, altname); err != nil {
		return err
	}
	// Do final link.
	if err := link(altname, name); err != nil {
		return err
	}
	return nil
}

// link the oldname to the newname.
// This method should be concurrency safe.
func link(oldname string, newname string) error {
	err := os.Remove(newname)
	// Ignore NotExist errors in case this is the first time the link is created.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Link(oldname, newname)
	if os.IsNotExist(err) || os.IsExist(err) {
		// Link was concurrently created or alt link was concurrently deleted or alt link never existed.
		return nil
	}

	return err
}

// write data to a new file described by the name with the provided mode.
func write(name string, mode os.FileMode, data []byte) error {
	f, err := os.Create(name)
	defer f.Close()
	if err != nil {
		return err
	}

	if err := f.Chmod(mode); err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return f.Sync()
}
