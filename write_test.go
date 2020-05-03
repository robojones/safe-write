package safe

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// checkContents compares the contents of a file with the wanted content.
func checkContents(t *testing.T, name string, want string) {
	got, err := ioutil.ReadFile(name)
	if err != nil {
		t.Error(err)
	} else {
		if string(got) != want {
			t.Errorf("Check contents of file %s got %q but want %q ", name, string(got), want)
		}
	}
}

// checkNotExist validates that a file does not exist.
func checkNotExist(t *testing.T, name string) {
	_, err := ioutil.ReadFile(name)
	if !os.IsNotExist(err) {
		t.Errorf("File %s should not exist", name)
	}
}

// clean removes a file or directory from the disk
func clean(t *testing.T, name string) {
	err := os.RemoveAll(name)
	if err != nil {
		t.Error(fmt.Errorf("Error during cleanup: %e", err))
	}
}

func createFile(t *testing.T, name string, contents string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(fmt.Errorf("create file for test: %s", err.Error()))
	}
	if contents != "" {
		if _, err := f.Write([]byte(contents)); err != nil {
			t.Fatal(fmt.Errorf("write file for test: %e", err))
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(fmt.Errorf("create file for test: %e", err))
	}
}

func createDir(t *testing.T, name string) {
	err := os.Mkdir(name, DefaultPerm)
	if err != nil {
		t.Fatal(fmt.Errorf("create directory for test: %e", err))
	}
}

func TestReadFile(t *testing.T) {
	t.Run("should read the testfile if it exists", func(t *testing.T) {
		createFile(t, "testfile", "some important data")
		defer clean(t, "testfile")
		got, err := ReadFile("testfile")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "some important data" {
			t.Errorf("ReadFile does not return the correct file contents. Want %q but got %q", "some important data", got)
		}
	})

	t.Run("should read testfile.1 if testfile does not exist", func(t *testing.T) {
		checkNotExist(t, "testfile")

		createFile(t, "testfile.1", "some important data")
		defer clean(t, "testfile.1")
		got, err := ReadFile("testfile")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "some important data" {
			t.Errorf("ReadFile does not return the correct file contents. Want %q but got %q", "some important data", got)
		}
	})

	t.Run("should return a NotExist error if neither testfile nor testfile.1 exists", func(t *testing.T) {
		checkNotExist(t, "testfile")
		checkNotExist(t, "testfile.1")

		_, err := ReadFile("testfile")
		if !os.IsNotExist(err) {
			t.Error(fmt.Errorf("expect NotExist error but got %e", err))
		}
	})

	t.Run("should automatically retry if testfile and testfile.1 do not exist and return when they are found within three intervals of 10ms", func(t *testing.T) {
		finishRead := make(chan bool)
		finishWrite := make(chan bool)

		go func() {
			checkNotExist(t, "testfile")

			got, err := ReadFile("testfile")
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != "some important data" {
				t.Errorf("ReadFile does not return the correct file contents. Want %q but got %q", "some important data", got)
			}

			finishRead <- true
		}()

		go func() {
			time.Sleep(10 * time.Millisecond)
			createFile(t, "testfile.1", "some important data")
			defer clean(t, "testfile.1")

			finishWrite <- true
		}()

		<-finishRead
		<-finishWrite
	})
}

func TestRemoveFile(t *testing.T) {
	t.Run("should not return an error if the files testfile and testfile.1 do not exist", func(t *testing.T) {
		err := RemoveFile("testfile")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("should return an error if testfile or testfile.1 can not be removed (in this case because they are a directory)", func(t *testing.T) {
		// create a directory with the name of the testfile so there is an error when using Remove
		createDir(t, "testfile")
		defer clean(t, "testfile")
		createFile(t, "testfile/somefile", "")

		err := RemoveFile("testfile")

		if err == nil {
			t.Errorf("expected an error but got nil")
		}
	})

	t.Run("should remove the files testfile and testfile.1", func(t *testing.T) {
		createFile(t, "testfile", "")
		createFile(t, "testfile.1", "")

		err := RemoveFile("testfile")
		if err != nil {
			t.Error(err)
		}

		checkNotExist(t, "testfile")
		checkNotExist(t, "testfile.1")
	})
}

func TestWriteFile(t *testing.T) {
	t.Run("should write the file to the disk and create the two links testfile and testfile.1", func(t *testing.T) {
		err := WriteFile("testfile", []byte("important contents"))
		if err != nil {
			t.Error(err)
		}
		defer clean(t, "testfile")
		defer clean(t, "testfile.1")

		checkContents(t, "testfile", "important contents")
		checkContents(t, "testfile", "important contents")
	})

	t.Run("should overwrite existing links when they exist from a previous write", func(t *testing.T) {
		err := WriteFile("testfile", []byte("data to be overwritten"))
		if err != nil {
			t.Error(err)
		}
		defer clean(t, "testfile")
		defer clean(t, "testfile.1")

		err = WriteFile("testfile", []byte("new data"))
		if err != nil {
			t.Error(err)
		}

		checkContents(t, "testfile", "new data")
		checkContents(t, "testfile.1", "new data")
	})

	t.Run("should return the error if the directory for the file does not exist", func(t *testing.T) {
		err := WriteFile("some/directory/which/does/not/exist/testfile", []byte("some data"))
		if !os.IsNotExist(err) {
			t.Error(fmt.Errorf("expect NotExist error but got %e", err))
		}
	})
}
