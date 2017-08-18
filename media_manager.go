package main

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	MediaPath string = "media"
)

var (
	paths = map[MediaType]string{
		MediaImage: "images",
		MediaVideo: "videos",
	}
)

type MediaManager struct {
	c  Configuration
	ca Cache
}

func NewMediaManager(cache Cache) *MediaManager {
	return &MediaManager{
		ca: cache,
	}
}

func (mm *MediaManager) Configure(configuration Configuration) error {
	mm.c = configuration

	return nil
}

func (mm *MediaManager) Save(file *File) (*Media, error) {
	var (
		e = filepath.Ext(file.Name)
		n = tempName("m-", e)
	)

	t, s, err := getPath(file.Type)
	if err != nil {
		return nil, err
	}

	p := filepath.Join(MediaPath, s, n)

	of, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	defer of.Close()
	if _, err = of.Write(file.Data); err != nil {
		return nil, err
	}

	hasher := sha256.New()
	hasher.Write(file.Data)

	mm.ca.Set(n, hex.EncodeToString(hasher.Sum(nil)))

	return &Media{
		Type: t,
		Path: p,
		Mime: file.Type,
	}, nil
}

func (mm *MediaManager) Delete(m *Media) error {
	return os.Remove(m.Path)
}

func (mm *MediaManager) PopulateEtagCache() {
	go providePopulatingFunc(MediaImage, sha256.New(), mm.ca)()
	go providePopulatingFunc(MediaVideo, sha256.New(), mm.ca)()
}

func providePopulatingFunc(t MediaType, h hash.Hash, c Cache) func() {
	return func() {
		p, ok := paths[t]
		if !ok {
			return
		}

		files, _ := ioutil.ReadDir(filepath.Join(MediaPath, p))
		for _, d := range files {
			data, err := ioutil.ReadFile(filepath.Join(MediaPath, p, d.Name()))
			if err != nil {
				continue
			}

			h.Write(data)
			c.Set(d.Name(), hex.EncodeToString(h.Sum(nil)))
			h.Reset()
		}
	}
}

func tempName(prefix, suffix string) string {
	return prefix + GenerateRandomString(16) + suffix
}

func getPath(s string) (MediaType, string, error) {

	var t MediaType
	{
		if strings.HasPrefix(s, "image") {
			t = MediaImage
		} else if strings.HasPrefix(s, "video") {
			t = MediaVideo
		} else {
			t = MediaOther
		}
	}

	if p, ok := paths[t]; ok {
		return t, p, nil
	}

	return t, "", ErrMediaNotSupported
}
