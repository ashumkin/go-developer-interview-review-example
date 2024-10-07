### Description
A service that returns sunrise time at the specified location.

It gets it either from two (imaginary) external  sources:
- http://sunrise.sunrise.io
- http://sun.ri.se

Return the first (of the two) response.

### Usage:

    POST /sunrise/at?location=...

or

    POST /sunrise/at?

    lat=...&lon=...
