# EVERYTHING Autonomous Agent (Magikarp)

![assets/shiny_Magikarp.webp](assets/shiny_Magikarp.webp)

## Description

The EVERYTHING Autonomous Agent, also known as Magikarp, is a Python FastAPI application that leverages (just) Dan's
data to deliver timely and accurate recommendations, simulate push notifications for a given date, and provide a chat
companion using AI and Large Language Models (LLMs).

## Key Features

- **AI and LLM Integration**: Utilizes Meta's Llama 3 model through [`ollama`](https://ollama.com/).
- **Chat Companion API**: Provides interactive chat functionality using the Magikarp model, which is based off Llama 3.
- **Personalized Recommendation API**: Delivers Dan prompt suggestions based on Dan's activities and preferences which Dan can
  select to interact with Magikarp.
- **Notification Simulation API**: Provides simulated push mobile notifications for Dan based on a provided date.

## Pre-requisites

- Python 3.12
- Poetry
- ollama (optional / `ollama` image used in docker commands to run the application)
- Docker

_Note: If you wish to download `ollama`, you can follow the instructions [here](https://ollama.com/). However, it is
easier to run the application through Docker by following the instructions in the next section._

## Quickstart

The easiest way to run this application is via docker.

### Docker 
To run the application, make sure you have your Docker daemon running (docker hub desktop app opened) and run the
following:

```shell
make docker-build && make docker-run
```

_Note: ollama uses ~ 4GB of disk space and can take ~10 minutes to install on initial run._


This command will do the following:

1. Run an `ollama` server, which runs Meta's Llama 3 model using the [Modelfile](./Modelfile) outline.
2. Run a FastAPI web server with APIs for interacting with Magikarp which has context on Dan's data.

The FastAPI webserver has two endpoints:

1. **Swagger endpoint:** The Swagger documentation, is available at [https://](http://127.0.0.1:8000/docs). This
   endpoint
   allows you to test out the APIs.
2. **Redoc endpoint (extra):** The Redoc documentation is a more stylish version of the Swagger documentation which does
   not allow for manual triggering of the APIs.

### Local developement

If you want to contribute to the development of the app, you must
[follow the instructions to download `ollama`](https://ollama.com/) and all other software dependencies outlined in the
pre-requisites section (except for Docker, obviously).

Once you have `ollama` installed on your system, you can run the following:

```shell
make model
```

This will create the Magikarp model based off Llama 3.

You can then run,

```shell
make run
```

This will run the FastAPI web server with reload available, allowing you to update the code and see your changes in real-time.


## How it Works

- The application runs a separate ollama server alongside the Python FastAPI webserver, enabling APIs for
  recommendations, notifications, and general AI prompting using Dan's data.
- A [`Modelfile`](./Modelfile) is used to create a pre-defined configuration for the Magikarp model based on the [instructions](./instructions.txt).
- User data would typically be stored locally on the user's device, with a short look-back period (e.g., 7 days). This
  data is ingested by the web server at runtime and used to formulate prompts to ollama.

## APIs

The application provides the following APIs:

### Recommendation API

- `/recommendations`: Provides predefined prompt suggestions for Dan to select and use to prompt Magikarp. These prompt
  suggestions are based off Dan's data.
- `/recommendations/suggest`: Generates a list of new prompt suggestions for Dan to use to prompt Magikarp. This API
  would
  ensure Dan's suggested prompts for Magikarp are updated throughout the day based on the time of day and his schedule.

### Notifications API

- `/notifications`: Sends simulated push-notifications to Dan based on predefined rules and his user data.

### Chat API

- `/chat`: Allows Dan to interact in a continuous conversation with Magikarp.