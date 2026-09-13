// Package user profile provides update operations for User aggregate
package user

//func (u *User) Update(user *User) error {
//	tracker := domain.NewDiffTracker()
//	tracker.TrackStringChange("email", u.Email, email)
//	return nil
//}
//
//func (u *User) ChangeEmail(tracker *domain.DiffTracker, email string) error {
//	if err := u.validateEmail(email); err != nil {
//		return commuser.ErrInvalidEmail
//	}
//
//	evt, err := eventsuser.NewUserUpdatedEvent(
//		u.ID,
//		u.AggregateName(),
//		eventsuser.UserUpdatedPayload{
//			Email: email,
//		},
//	)
//	if err != nil {
//		return commuser.ErrInvalidEmail
//	}
//
//	if err := u.RaiseEvent(evt); err != nil {
//		return err
//	}
//	return nil
//}
//
//func (u *User) validateEmail(email string) error {
//	if email == "" {
//		return fmt.Errorf("email cannot be empty")
//	}
//
//	addr, err := mail.ParseAddress(email)
//	if err != nil {
//		return fmt.Errorf("invalid email format: %w", err)
//	}
//
//	if len(addr.Address) > 128 {
//		return fmt.Errorf("email address too long (max 128 characters)")
//	}
//
//	return nil
//}
