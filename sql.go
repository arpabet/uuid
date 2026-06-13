/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid

import (
	"database/sql/driver"
	"github.com/pkg/errors"
)

/**
Value implements the database/sql/driver.Valuer interface.

The UUID is stored as its canonical 36 character string representation, which
is accepted by the native uuid type in PostgreSQL and as CHAR/VARCHAR elsewhere.
*/

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

/**
Scan implements the database/sql.Scanner interface.

Accepts a nil value, the canonical string form (any form supported by Parse),
a textual byte slice, or a raw 16 byte binary representation.
*/

func (u *UUID) Scan(src interface{}) error {

	switch v := src.(type) {

	case nil:
		return nil

	case string:
		if v == "" {
			return nil
		}
		parsed, err := Parse(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			return nil
		}
		// a raw binary UUID is exactly 16 bytes; every textual form is longer
		if len(v) == 16 {
			return u.UnmarshalBinary(v)
		}
		parsed, err := ParseBytes(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	default:
		return errors.Errorf("uuid: cannot scan type %T into UUID", src)
	}
}
