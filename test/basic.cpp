#include <iostream>
#include <mpi.h>

int main(int argc, char** argv) {
    // Initialize MPI environment
    MPI_Init(&argc, &argv);

    int rank, size;
    MPI_Comm_rank(MPI_COMM_WORLD, &rank); // Get the process ID (rank)
    MPI_Comm_size(MPI_COMM_WORLD, &size); // Get the total number of processes

    // Define the total number of elements to process
    const int total_elements = 10;

    // Calculate the number of elements to process per rank
    const int elements_per_rank = total_elements / size;
    
    // Perform computation in parallel
    int start_idx = rank * elements_per_rank;
    int end_idx = (rank + 1) * elements_per_rank;

    // In case total_elements is not divisible evenly by the number of processes,
    // the last rank will handle the remaining elements.
    if (rank == size - 1) {
        end_idx = total_elements;
    }

    // Perform some computation on the elements
    for (int i = start_idx; i < end_idx; ++i) {
        // Example: Square each element
        int result = i * i;

        // Print the result for each rank
        std::cout << "Rank " << rank << ": Element " << i << " squared is " << result << std::endl;
    }

    // Finalize MPI environment
    MPI_Finalize();
    return 0;
}
