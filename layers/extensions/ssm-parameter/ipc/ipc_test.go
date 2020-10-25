package ipc_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"ssm-parameter/ipc"
	"ssm-parameter/ipc/mock"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Run("responds to a request and returns the parameter", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ssmAPI := mock.NewMockSSMAPI(ctrl)
		ssmAPI.EXPECT().GetParameter(&ssm.GetParameterInput{
			Name: aws.String("extension-parameter"),
		}).Return(&ssm.GetParameterOutput{
			Parameter: &ssm.Parameter{Value: aws.String("parameter-value")}}, nil).Times(1)

		server := ipc.New("3000", ssmAPI)

		go server.Start(ctx)

		resp, err := http.Get("http://localhost:3000")
		respBuf, _ := ioutil.ReadAll(resp.Body)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "{\"body\":\"parameter-value\"}", string(respBuf))
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("handles errors when invoking ssm api", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ssmAPI := mock.NewMockSSMAPI(ctrl)
		ssmAPI.EXPECT().GetParameter(&ssm.GetParameterInput{
			Name: aws.String("extension-parameter"),
		}).Return(nil, errors.New("boom")).Times(1)

		server := ipc.New("3000", ssmAPI)

		go server.Start(ctx)

		resp, err := http.Get("http://localhost:3000")
		respBuf, _ := ioutil.ReadAll(resp.Body)

		require.NoError(t, err)
		require.Equal(t, string(respBuf), "{\"body\":\"boom\"}")
	})

	t.Run("subsequent results are provided from cache", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ssmAPI := mock.NewMockSSMAPI(ctrl)
		ssmAPI.EXPECT().GetParameter(&ssm.GetParameterInput{
			Name: aws.String("extension-parameter"),
		}).Return(&ssm.GetParameterOutput{
			Parameter: &ssm.Parameter{Value: aws.String("parameter-value")}}, nil).Times(1)

		server := ipc.New("3000", ssmAPI)

		go server.Start(ctx)

		resp, err := http.Get("http://localhost:3000")
		respBuf, _ := ioutil.ReadAll(resp.Body)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "{\"body\":\"parameter-value\"}", string(respBuf))
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		secondResp, err := http.Get("http://localhost:3000")
		secondRespBuf, _ := ioutil.ReadAll(secondResp.Body)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, secondResp.StatusCode)
		require.Equal(t, "{\"body\":\"parameter-value\"}", string(secondRespBuf))
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})
}
