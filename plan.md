

MVP:

Create a file through websocket connection between magikarp and agent.



workflow:

1. run a cli command to create a new project, provide project outline and 
select the model you want to interact with the project.
then api is triggered to setup the model, create the project directory and add basic 
files (README.md, .gitignore, etc.). this is printed to the user. ask if ready to proceed, agent commands will require
approval before execution [y/N].
2. if [y] the model is prompted then:
   1. ollama server is started
   2. /chat api is hit with initial SNAPSHOT 1 of project structure and file contents
   3. model is prompted to interact with the project given this basic info (SNAPSHOT 1)
3. the response from the model is sent to the agent, which triggers the appropriate API to interact with the project
4. the agent then sends the updated SNAPSHOT 2 of the project structure and file contents to the model 
5. the model is prompted to interact with the project given this updated info (SNAPSHOT 2)
6. the response from the model is sent to the agent, which triggers the appropriate API to interact with the project
7. the agent then sends the updated SNAPSHOT 3 of the project structure and file contents to the model
8. and so on...


and the agent triggers APIs to interact with the project


websocket

on one end, model is prompted to interact with project
on the other end, agent triggers APIs to interact with project


Model configuration:

You are an autonomous agent tasked with building a fully working coding project.

The initial prompt you receive will be of a project outline, including goals that you need to achieve through iteration.
The initial prompt will also be accompanied by a "SNAPSHOT 1" of the project directory, and each file's contents.

Your responses will be connected to a websocket wherein the agent on the other end will trigger APIs to interact with
the project.

Here are the APIs you can trigger:

```json
{
  "apis": [
    {"key": "write", "description": "Write/update a file in the project"},
    {"key": "delete", "description": "Delete a file in the project"},
    {"key": "read", "description": "Read a file in the project"},
    {"key": "read-all", "description": "Read all files in the project"},
    {"key": "list", "description": "List all files in the project (using tree)"},
    {"key": "command", "description": "Run a CLI command"}
  ]
}
```

To trigger these APIs, your responses should be of the format:

```json
{
  "key": "write",
  "file_name": "file.txt",
  "content": "Hello, World!",
  "update": false
}
```

You can also trigger the creation of multiple files in a single request:

```json
[
  {
  "api": "/write",
  "files": ["file1.txt", "file2.txt"]
  },
  {
    "api": "/write",
    "file_name": "file1.txt",
    "content": "Hello, World!"
  }
]
```

Following user specific requirements for the project, create files in the project directory and run terminal commands to
achieve the project goals and resolve issues.

Each response by the user will detail a SNAPSHOT of the project directory, and each file's contents. React to the
output and iterate on the commands by triggering the next API.


---

APIs to:

1. Write/update file in a project
2. Delete a file in a project
3. Read a file in a project
   1. Read all files in a project
4. List all files in a project (using tree)
5. Run commands and review output/iterate on commands