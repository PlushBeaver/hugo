// Copyright 2018 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugofs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestRootMappingFsMount(t *testing.T) {
	assert := require.New(t)
	fs := afero.NewMemMapFs()

	// from = themename to := b.p.AbsPathify(filepath.Join(b.p.ThemesDir, theme.Name, themeFolder))
	testfile := "test.txt"

	assert.NoError(afero.WriteFile(fs, filepath.Join("themes/a/myblogcontent", testfile), []byte("some content"), 0755))

	bfs := NewBasePathRealFilenameFs(afero.NewBasePathFs(fs, "themes/a").(*afero.BasePathFs))
	rm := RootMapping{
		From: "blog",
		To:   "myblogcontent",
		Lang: "no",
	}

	rfs, err := NewRootMappingFs(bfs, rm)
	assert.NoError(err)

	blog, err := rfs.Stat("blog")
	assert.NoError(err)
	assert.Equal("myblogcontent", blog.(RealFilenameInfo).RealFilename())
	assert.Equal("no", blog.(LangProvider).Lang())

	//bf, err := blog.(FileOpener).Open()
	//assert.NoError(err)
	//defer bf.Close()

	dirs, err := afero.ReadDir(rfs, "blog")
	assert.NoError(err)
	assert.Equal(1, len(dirs))
	dir := dirs[0]
	assert.Equal(testfile, dir.Name())

	rfi, ok := dir.(RealFilenameInfo)
	assert.True(ok)
	assert.Equal("themes/a/myblogcontent/test.txt", rfi.RealFilename())

	pfi, ok := dir.(FilePathPather)
	assert.True(ok)
	assert.Equal("blog/test.txt", pfi.Path())

}

func TestRootMappingFsRealName(t *testing.T) {
	assert := require.New(t)
	fs := afero.NewMemMapFs()

	rfs, err := NewRootMappingFsFromFromTo(fs, "f1", "f1t", "f2", "f2t")
	assert.NoError(err)

	name, _, _ := rfs.realNameAndRoot(filepath.Join("f1", "foo", "file.txt"))
	assert.Equal(filepath.FromSlash("f1t/foo/file.txt"), name)

}

func TestRootMappingFsDirnames(t *testing.T) {
	assert := require.New(t)
	fs := afero.NewMemMapFs()

	testfile := "myfile.txt"
	assert.NoError(fs.Mkdir("f1t", 0755))
	assert.NoError(fs.Mkdir("f2t", 0755))
	assert.NoError(fs.Mkdir("f3t", 0755))
	assert.NoError(afero.WriteFile(fs, filepath.Join("f2t", testfile), []byte("some content"), 0755))

	rfs, err := NewRootMappingFsFromFromTo(fs, "bf1", "f1t", "cf2", "f2t", "af3", "f3t")
	assert.NoError(err)

	fif, err := rfs.Stat(filepath.Join("cf2", testfile))
	assert.NoError(err)
	assert.Equal("myfile.txt", fif.Name())
	assert.Equal(filepath.FromSlash("f2t/myfile.txt"), fif.(RealFilenameInfo).RealFilename())

	root, err := rfs.Open(filepathSeparator)
	assert.NoError(err)

	dirnames, err := root.Readdirnames(-1)
	assert.NoError(err)
	assert.Equal([]string{"bf1", "cf2", "af3"}, dirnames)

}

func TestRootMappingFsOs(t *testing.T) {
	assert := require.New(t)
	fs := afero.NewOsFs()

	d, err := ioutil.TempDir("", "hugo-root-mapping")
	assert.NoError(err)
	defer func() {
		os.RemoveAll(d)
	}()

	testfile := "myfile.txt"
	assert.NoError(fs.Mkdir(filepath.Join(d, "f1t"), 0755))
	assert.NoError(fs.Mkdir(filepath.Join(d, "f2t"), 0755))
	assert.NoError(fs.Mkdir(filepath.Join(d, "f3t"), 0755))
	assert.NoError(afero.WriteFile(fs, filepath.Join(d, "f2t", testfile), []byte("some content"), 0755))

	rfs, err := NewRootMappingFsFromFromTo(fs, "bf1", filepath.Join(d, "f1t"), "cf2", filepath.Join(d, "f2t"), "af3", filepath.Join(d, "f3t"))
	assert.NoError(err)

	fif, err := rfs.Stat(filepath.Join("cf2", testfile))
	assert.NoError(err)
	assert.Equal("myfile.txt", fif.Name())

	root, err := rfs.Open(filepathSeparator)
	assert.NoError(err)

	dirnames, err := root.Readdirnames(-1)
	assert.NoError(err)
	assert.Equal([]string{"bf1", "cf2", "af3"}, dirnames)

}
