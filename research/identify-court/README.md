# Tennis Court 3D Model Generator

This script generates a 3D model of a tennis court in GLB format with accurate dimensions.

## Features

- Creates a tennis court model with true-to-life dimensions
- Includes court lines, net posts, net, and center tape
- Exports the model in GLB format, which is a binary version of GLTF that contains all data in a single file

## Requirements

- Python 3.6+
- Dependencies listed in `requirements.txt`

## Installation

### Option 1: Local Installation

1. Clone this repository
2. Install the required dependencies using either `pip` or `uv`:

Using `pip`:

```bash
pip install -r requirements.txt
```

Using `uv` (faster alternative to pip):

```bash
uv pip install -r requirements.txt
```

### Option 2: Docker Installation

1. Clone this repository
2. Build the Docker image:

```bash
docker build -t tennis-court-generator .
```

### Option 3: Docker Compose Installation

1. Clone this repository
2. Create an output directory:

```bash
mkdir -p output
```

## Usage

### Local Usage

Run the script to generate the tennis court model:

```bash
python generate_court_model.py
```

This will create a file named `tennis_court.glb` in the current directory.

### Docker Usage

Run the Docker container to generate the tennis court model:

```bash
docker run -v $(pwd):/app/output tennis-court-generator
```

This will create a file named `tennis_court.glb` in your current directory.

### Docker Compose Usage

Run the Docker container using docker-compose:

```bash
docker-compose up
```

This will create a file named `tennis_court.glb` in the `output` directory.

## Court Dimensions

The model uses the following standard tennis court dimensions:

- Court length: 78ft (23.77m)
- Singles court width: 27ft (8.23m)
- Doubles court width: 36ft (10.97m)
- Service line distance: 21ft (6.4m) from the net
- Net height at posts: 42" (106.7 cm)
- Net height at center: 36" (91.4 cm)
- Net width: 42' (12.8 m)
- Net post diameter: ~2.75" (7 cm)
- Line width: 2" (5.1 cm)
- Baseline width: 4" (10.2 cm)

## Viewing the Model

You can view the generated GLB file using:

- [Three.js Editor](https://threejs.org/editor/)
- [Babylon.js Sandbox](https://sandbox.babylonjs.com/)
- [Microsoft 3D Viewer](https://www.microsoft.com/en-us/p/3d-viewer/9nblggh42ths) (Windows)
- [Blender](https://www.blender.org/) (with GLTF import plugin)
- [Google Model Viewer](https://modelviewer.dev/) (for web applications)
