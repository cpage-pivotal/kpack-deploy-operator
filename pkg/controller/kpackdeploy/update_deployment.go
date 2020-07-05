package kpackdeploy

import "regexp"

var imageTagRegex = regexp.MustCompile("image: .+")

func updateDeployment(latestImage, oldContent string) (newContent string, err error) {
	return imageTagRegex.ReplaceAllString(oldContent, "image: "+latestImage), nil
}
