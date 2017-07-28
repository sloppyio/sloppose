# Sloppose

Small tool to convert docker-compose files to sloppy.io compatible ones.

Supports docker-compose versions **2** and **3**.

## Usage

`sloppose [command]`

**Commands**:
* `convert [options] [files]`
    * Example: `sloppose convert -o outFile.yml -projectname example`

## Configuration

**Projectname**:
* can be set with `COMPOSE_PROJECT_NAME` environment variable or with parameter as seen above.
* defaults to current working dir
