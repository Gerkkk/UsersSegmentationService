package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	segv1 "main/protos/gen/go/segmentation"
	"main/tests/suite"
	"testing"
)

func TestAllUseCases(t *testing.T) {
	ctx, st := suite.New(t)

	segId := "MAIL_VOICE_MESSAGES"
	segDescription := "Adding voice messages to mail"

	respReg, err := st.AuthClient.CreateSegment(ctx, &segv1.CreateSegmentRequest{
		Id:          segId,
		Description: segDescription,
	})
	require.NoError(t, err)
	assert.Equal(t, respReg.Id, segId)

	respReg, err = st.AuthClient.CreateSegment(ctx, &segv1.CreateSegmentRequest{
		Id:          segId,
		Description: segDescription,
	})

	require.Error(t, err)

	newSegDescription := "Adding cool voice messages to mail"

	updSeg, err := st.AuthClient.UpdateSegment(ctx, &segv1.UpdateSegmentRequest{
		Id:             segId,
		NewDescription: &newSegDescription,
	})
	require.NoError(t, err)
	assert.Equal(t, updSeg.Id, segId)

	infoSeg, err := st.AuthClient.GetSegmentInfo(ctx, &segv1.GetSegmentInfoRequest{
		Id: segId,
	})
	require.NoError(t, err)
	assert.Equal(t, infoSeg.Description, newSegDescription)

	perc := "100"
	distrSeg, err := st.AuthClient.DistributeSegment(ctx, &segv1.DistributeSegmentRequest{
		Id:              segId,
		UsersPercentage: perc,
	})
	require.NoError(t, err)
	assert.Equal(t, distrSeg.Id, segId)

	var userId int64
	userId = 1
	_, err = st.AuthClient.GetUserSegments(ctx, &segv1.GetUserSegmentsRequest{
		Id: userId,
	})
	require.Error(t, err)

	delSeg, err := st.AuthClient.DeleteSegment(ctx, &segv1.DeleteSegmentRequest{
		Id: segId,
	})
	require.NoError(t, err)
	assert.Equal(t, delSeg.Id, segId)
}
