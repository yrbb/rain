package orm

func (s *Session) Begin() error {
	if s.txID > 0 {
		return nil
	}

	if s.tx != nil {
		return ErrTransExist
	}

	tx, err := s.orm.db.Begin()
	if err != nil {
		return err
	}

	s.tx = tx

	return nil
}

func (s *Session) Rollback() error {
	return s.transaction("rollback")
}

func (s *Session) Commit() error {
	if err := s.transaction("commit"); err != nil {
		return err
	}

	return nil
}

func (s *Session) transaction(t string) error {
	if s.txID > 0 {
		return nil
	}

	if s.tx == nil {
		return ErrTransNotExist
	}

	var err error
	if t == "rollback" {
		err = s.tx.Rollback()
	} else {
		err = s.tx.Commit()
	}

	s.tx = nil

	return err
}
