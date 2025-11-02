package booking

import (
	"fmt"

	"bk-concerts/applications/participant"
	"bk-concerts/logger"
)

// GetAllParticipantByBookingID retrieves all participants for a given booking ID.
func GetAllParticipantByBookingID(bookingID string) ([]*participant.Participant, error) {
	logger.Log.Info(fmt.Sprintf("[get-all-participants-uc] Fetching participants for bookingID: %s", bookingID))

	// Fetch booking first
	bk, err := GetBooking(bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-participants-uc] Failed to fetch booking: %v", err))
		return nil, fmt.Errorf("failed to fetch booking: %w", err)
	}

	if len(bk.ParticipantIDs) == 0 {
		logger.Log.Info(fmt.Sprintf("[get-all-participants-uc] No participants linked with bookingID: %s", bookingID))
		return []*participant.Participant{}, nil
	}

	// Retrieve each participant
	var participants []*participant.Participant
	for _, pid := range bk.ParticipantIDs {
		pt, err := participant.GetParticipant(pid)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-participants-uc] Failed to fetch participantID: %s, error: %v", pid, err))
			return nil, fmt.Errorf("failed to fetch participantID %s: %w", pid, err)
		}
		participants = append(participants, pt)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-participants-uc] Successfully retrieved %d participants for bookingID: %s", len(participants), bookingID))
	return participants, nil
}
