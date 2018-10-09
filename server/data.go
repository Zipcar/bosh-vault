package server

import (
	"github.com/labstack/echo"
	vcfcsTypes "github.com/zipcar/vault-cfcs/types"
	"io/ioutil"
	"net/http"
)

func dataPostHandler(ctx echo.Context) error {
	context := ctx.(*ConfigurationContext)
	requestBody, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err))
		return err
	}

	context.Log.Debugf("request: %s", requestBody)

	credentialRequest, err := vcfcsTypes.ParseGenericCredentialRequest(requestBody)
	if err != nil {
		context.Log.Error("request error: ", err)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, err.Error()))
		return err
	}

	
	ok := credentialRequest.Validate()
	if !ok {
		credentialType := credentialRequest.CredentialType()
		context.Log.Error("invalid credential request for ", credentialType)
		ctx.Error(echo.NewHTTPError(http.StatusBadRequest, "invalid credential request for ", credentialType))
		return err
	}

	context.Log.Debugf("attempting to generate %s", credentialType)

	err = credentialRequest.Generate()
	if err != nil {
		context.Log.Error(err)
		ctx.Error(echo.NewHTTPError(http.StatusInternalServerError, "problem generating ", credentialType, err))
	}

	return ctx.JSON(http.StatusOK, &map[string]interface{}{
		"status":      http.StatusOK,
		"status_text": http.StatusText(http.StatusOK),
	})
}
