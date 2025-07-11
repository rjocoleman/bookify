// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.906
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func SetupPage() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = SetupPageWithError("").Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func SetupPageWithError(errorMsg string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>Setup - Bookify</title><script src=\"https://unpkg.com/htmx.org@1.9.10\"></script><script src=\"https://cdn.tailwindcss.com\"></script></head><body class=\"bg-gray-100 min-h-screen\"><div class=\"max-w-md mx-auto pt-8\"><div class=\"bg-white rounded-lg shadow p-6\"><h1 class=\"text-2xl font-bold mb-6 text-center\">Setup Bookify</h1>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if errorMsg != "" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<div class=\"mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(errorMsg)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/setup.templ`, Line: 23, Col: 17}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<p class=\"text-gray-600 mb-6 text-sm\">Connect your Google Drive account to start converting and uploading EPUB files.</p><div class=\"mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg\"><p class=\"text-sm text-blue-800 mb-2\"><strong>How it works:</strong></p><ol class=\"text-sm text-blue-700 list-decimal list-inside space-y-1\"><li>Enter a name for this account</li><li>Enter your Google Drive folder ID</li><li>Authorize Bookify to access your Google Drive</li><li>Your files will be uploaded to your own Google Drive storage</li></ol></div><form hx-post=\"/setup\" hx-target=\"body\" class=\"space-y-4\"><div><label class=\"block text-sm font-medium text-gray-700 mb-1\">Account Name</label> <input name=\"name\" type=\"text\" required class=\"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500\" placeholder=\"e.g., Personal Account\"></div><div><label class=\"block text-sm font-medium text-gray-700 mb-1\">Google Drive Folder ID</label> <input name=\"folder_id\" type=\"text\" required class=\"w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500\" placeholder=\"e.g., 1A2B3C4D5E6F7G8H9I0J\"><p class=\"text-xs text-gray-500 mt-1\">The ID is in the folder URL: drive.google.com/drive/folders/<strong>[FOLDER_ID]</strong></p></div><button type=\"submit\" class=\"w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition duration-200\">Authorize with Google</button></form><div class=\"mt-4 text-center\"><a href=\"/\" class=\"text-sm text-gray-600 hover:underline\">Back to Home</a></div></div></div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
