### Work List
- [x] Write a docker client that will be used by the tools to access docker end points through the SDK.
- [ ] Write a tool adapter that takes the name of the tool and its input so that we can do confirmation.
- [ ] Make a tool interface and decide what docker functionalities to include in the tools.
- [ ] Implement context initialization that will get included in the system prompt to indicate the current state of the user.
- [ ] Maybe make an initialization function that takes the config parameters and creates a Genkit client (may require initialization of the docker client as well).
- [ ] Implement persistence memory with sessions (might require work to do, not sure).
- [ ] Integrate the tools with `user_prompt` and Agent Flow.
- [ ] Adjust the orchestrator to stream thoughts and responses (maybe use websocket later for web? and for TUI use the Genkit package streaming they mentioned).

### References:
- https://genkit.dev/docs/go/flows/#using-iterator-based-streaming
- https://genkit.dev/docs/go/tool-calling/#explicitly-handling-tool-calls
- https://genkit.dev/docs/go/chat/
- https://pkg.go.dev/github.com/moby/moby/client#section-readme
