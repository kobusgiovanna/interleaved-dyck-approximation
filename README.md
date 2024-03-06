
# A Better Approximation for Interleaved Dyck Reachability


This repository contains an implementation of the algorithms described in "A Better Approximation for Interleaved Dyck Reachability" by Giovanna Kobus Conrado and Andreas Pavlogiannis.


The algorithm is a modified version of a code created and written by Andreas Pavlogiannis, Jaco van de Pol and dam Husted Kjelstrøm. The benchmarks and parts of the code are from the implementation of "Mutual Refinements of Context-Free Language Reachability" by Shuo Ding and Qirun Zhang.

**Quick start:** To run the code, go to ```src/main/``` and run 

```python3 run.py``` 

in the terminal. Make sure you have Go and Python 3 installed.


## Structure

All code is stored in the ```src/main/``` folder.

```run.py``` contain the function to run the two sets of benchmarks: taint and valueflow. Benchmarks are located in their respective folders.

```main.go``` contains the main function that runs the algorithms described.

The remaining files are auxiliary files defining data structures and tests.