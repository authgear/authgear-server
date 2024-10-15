package facerecognition

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var NoMatchingFaceFound apierrors.Kind = apierrors.Invalid.WithReason("NoMatchingFaceFound")
var SpoofedImageDetected apierrors.Kind = apierrors.Invalid.WithReason("SpoofedImageDetected")
var FaceTooSmall apierrors.Kind = apierrors.Invalid.WithReason("FaceTooSmall")
var FaceRotated apierrors.Kind = apierrors.Invalid.WithReason("FaceRotated")
var FaceEdgesNotVisible apierrors.Kind = apierrors.Invalid.WithReason("FaceEdgesNotVisible")
var FaceCovered apierrors.Kind = apierrors.Invalid.WithReason("FaceCovered")
var FaceTooClose apierrors.Kind = apierrors.Invalid.WithReason("FaceTooClose")
var FaceCropped apierrors.Kind = apierrors.Invalid.WithReason("FaceCropped")
var LowImageQuality apierrors.Kind = apierrors.Invalid.WithReason("LowImageQuality")
var MultipleFaces apierrors.Kind = apierrors.Invalid.WithReason("MultipleFaces")
var NoFaceDetected apierrors.Kind = apierrors.Invalid.WithReason("NoFaceDetected")
