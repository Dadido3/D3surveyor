# D3surveyor ![building and testing state](https://github.com/Dadido3/D3surveyor/actions/workflows/build-test.yml/badge.svg?branch=master)

This software can be used to create 3-dimensional meshes/points from a series of measurements.
All that is needed is a list of points (e.g. corners of a room), distance measurements between these points and optionally some photos describing the angle between several points.

With enough measurements and constraints, the software finds the point positions that best fit the given measurements.
Ideally, the points then correspond to the real geometry of the object to be measured.
Additional measurements can be added to increase accuracy.

[Click here to run the current release in the browser!](https://dadido3.github.io/D3surveyor/)

> :warning: This is a proof-of-concept that works to some extent, but still has many rough edges and limitations.

![Image showing camera settings and a list of taken photos](/images/example-camera.png) ![Image showing the points mapping editor](/images/example-camera-photo.png)

## Usage

Let's say you want to create a 3D model of a room and need its exact geometry.
You need

- something to measure distances (e.g. a laser rangefinder), and optionally
- something to take wide angle photos with (e.g. a phone).

Here are the basic steps of how to achieve good (or any) results:

1. Create a new camera object and set its `Angle of view` to match coarsely the long side angle of view of your images you gonna take.
2. Lock the `Angle of view` parameter.
3. Add new photos to the camera object.
   Either import previously taken images, or directly capture new ones on your phone.
   The images should contain as many points of interest as possible, and they should be taken from different positions and perspectives.
   Also, don't mix different angle of views (don't cut images, don't change the zoom level).
   If you want to use images with different angle of views, create new camera objects for these.
4. Create points and name them accordingly, like `Room NWT` for the north west top corner of the room.
5. Get into the image edit mode of every photo and map all points to every image.
   Double click to add a flag ("point mapping"), and single click to change which point it maps to.
6. Lock all the position and rotation parameters of **one** camera to prevent the points from floating around into infinity.
   Alternatively, you can lock the position of some points to some known coordinates.
7. Create a rangefinder object and add measurements to it.
8. Press the "reload" icon in the sidebar to let the software recalculate the points.
9. Press the "save" icon to save the current state.
   Press the "export" icon to export an Wavefront OBJ file that contains the points.

## Useful information

- Keep the `Angle of view` parameter of cameras locked, because if the software can't find a good matching solution it will find the trivial solution.
  That is all cameras floating into infinity while the the angle of view approaches 0.
  Once there is a good solution, you can unlock it to let the optimizer find a better `Angle of view`.
- If you already have a good network of points and want to add an additional photo, you only need to add 3 point mappings (flags) to let the optimizer find the photo's origin and orientation.
  Once the photo is correctly aligned in the 3D space, the software will show suggested point mappings that can be confirmed by double clicking on them.

## Compiling

To compile the software yourself, all you need is an installation of [Go](https://golang.org).
There are at least two ways to compile and run the application:

### Development server

If you want to quickly test changes, you just have to run the development server.
If you use Visual Studio Code, you only have to start the debugger (F5).
Otherwise run

``` bash
cd ./scripts
go run ./devserver.go
```

Once the server is running, you can connect via [http://localhost:8875](http://localhost:8875).
The server will compile the project every time you reload the browser.

### Generate distribution

If you want to host it with any available webserver, just run

``` bash
go run ./scripts/dist.go
```

This will create a `dist` directory that contains all files needed to serve the application from a webserver.

> :information_source: Opening `index.html` directly in a browser doesn't work. You need to serve it via a webserver.

If the application is not served from the root directory of your domain, you need to pass the path prefix to `dist.go`.
Assuming you want to access the application via `example.com/some/sub/folder`, you have to use:

``` bash
go run ./scripts/dist.go -urlpathprefix /some/sub/folder
```

## Issues

- Amount of possible images is limited by how much RAM the browser allows.
  All images have to be stored in RAM.
- Reloading the page will reset all progress, so save regularly.
  Navigation back and forth works, though.
- The UI isn't optimal for working with a huge amount of points, cameras or measurements.
  Also, the optimizer isn't best suited for this kind of problem.

## Future

There are many possible features that could be added in the future:

- Other coordinate systems (e.g. WGS 84).
- Use of the GPS metadata from images.
- (More) constraints like parallelism, alignment to primary axes.
- Tags to filter objects.
- Optimizer improvements.
- Suggest points or other parameters when creating new measurements.
- Documentation
