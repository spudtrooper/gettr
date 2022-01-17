package todo

import "log"

func SkipErr(while string, err error) {
	log.Printf("TODO: Skipping error while %s: %v", while, err)
}

func NilError(while string) error {
	log.Printf("TODO: Skipping returning nil while %s", while)
	return nil
}
