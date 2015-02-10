package handlers

import ()

// HomeHandler temp landing. FIXME
func HomeHandler(response Responser, playload Payloader) {
	response.Write([]byte("Hello Developer"))
}
