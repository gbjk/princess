A naive weighed dummy webxg backend which responds to requests in about 400ms,
and tells the proxy it's ready in about 4000ms.

Both response times are weighted toward the lower end, with a long tail end towards rare slow responses.
