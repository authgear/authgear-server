package handlers

import ()

// HomeHandler temp landing. FIXME
func HomeHandler(response Responser, playload Payload) {
	response.Write([]byte("Hello Developer"))
}
