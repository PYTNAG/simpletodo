package validation

import (
	"fmt"
	"regexp"
)

const (
	InfUpperBound = -1
	InfLowerBound = -1
)

var (
	isValidUsername = regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString
)

func ValidateStringLength(str string, min, max int) error {
	length := len(str)

	if (min != InfLowerBound && length < min) || (max != InfUpperBound && length > max) {
		if max == InfUpperBound {
			return fmt.Errorf("must contain at least %d characters", min)
		}

		if min == InfLowerBound {
			return fmt.Errorf("must contain at most %d characters", max)
		}

		return fmt.Errorf("must contain from %d to %d characters", min, max)
	}

	return nil
}

func ValidateInteger(integer, min, max int) error {
	if integer < min || integer > max {
		return fmt.Errorf("must be in range from %d to %d", min, max)
	}

	return nil
}

func ValidateUsername(username string) error {
	if err := ValidateStringLength(username, 1, InfUpperBound); err != nil {
		return err
	}

	if !isValidUsername(username) {
		return fmt.Errorf("must contain only letters, digits or underscore")
	}

	return nil
}

func ValidatePassword(password string) error {
	return ValidateStringLength(password, 6, InfUpperBound)
}

func ValidateId(id int32) error {
	if err := ValidateInteger(int(id), 1, (1<<32)-1); err != nil {
		return err
	}

	return nil
}

func ValidateStringParam(str string) error {
	if err := ValidateStringLength(str, 1, InfUpperBound); err != nil {
		return err
	}

	return nil
}
