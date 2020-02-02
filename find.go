package watchman

import (
	"github.com/yookoala/realpath"
)

// Find is the return object of the Find call
type Find struct {
	Clock string `bser:"clock"`
	Files []File `bser:"files"`
}

// File represents a file on the filesystem
type File struct {
	Cclock string `bser:"cclock"`
	Ctime  int    `bser:"ctime"`
	Dev    int    `bser:"dev"`
	Exists bool   `bser:"exists"`
	Gid    int    `bser:"gid"`
	Ino    int    `bser:"ino"`
	Mode   int    `bser:"mode"`
	Mtime  int    `bser:"mtime"`
	Name   string `bser:"name"`
	New    bool   `bser:"new"`
	Nlink  int    `bser:"nlink"`
	Oclock string `bser:"oclock"`
	Size   int    `bser:"size"`
	UID    int    `bser:"uid"`
}

// Find finds all files that match the optional list of patterns under the specified dir. If no patterns were specified, all files are returned.
func (c *Client) Find(path string, patterns ...string) (*Find, error) {
	path, err := realpath.Realpath(path)
	if err != nil {
		return nil, err
	}

	var data struct {
		base
		Find
	}

	args := make([]interface{}, len(patterns)+2)
	args[0] = "find"
	args[1] = path
	for i := 0; i < len(patterns); i++ {
		args[i+2] = patterns[i]
	}

	if err := c.Send(&data, args...); err != nil {
		return nil, err
	}

	if data.Error != "" {
		return nil, data.Error
	}

	return &data.Find, nil
}
