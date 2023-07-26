#include <iostream>
#include <mpi.h>
#include <thread>
#include <vector>
#include <chrono>

using namespace std;

int initialNodeConfig[9][3] = {
        {2, 3, 345},
        {12, 45, 234},
        {35, 45, 533},
        {1, 38, 234},
        {4, 5, 50},
        {50, 3, 98},
        {22, 2, 144},
        {34, 33, 233},
        {11, 39, 235}
};
int x, y, elevel;

// Function to be executed by each thread
void count(int rank) {
    x = initialNodeConfig[rank][0];
    y = initialNodeConfig[rank][1];
    elevel = initialNodeConfig[rank][2];

    while(elevel > 0) {
        cout << "Thread: " << rank << " energy level is: " << elevel<< endl;
        // Perform some computation here (e.g., use rank for different tasks)
        // Sleep 1 second
        this_thread::sleep_for(chrono::milliseconds(1000));
        elevel --;
    }
}

void threadFunction2(int rank) {
    std::cout << "Thread f2 " << rank << " is running." << std::endl;
    // Perform some computation here (e.g., use rank for different tasks)
}

int main(int argc, char** argv) {
    // Initialize MPI environment
    MPI_Init_thread(&argc, &argv, MPI_THREAD_SERIALIZED, nullptr);

    int rank, size;
    MPI_Comm_rank(MPI_COMM_WORLD, &rank); // Get the process ID (rank)
    MPI_Comm_size(MPI_COMM_WORLD, &size); // Get the total number of processes

    thread t1(count, rank);
    thread t2(threadFunction2, rank);

    t1.join();
    t2.join();
    // Finalize MPI environment
    MPI_Finalize();
    return 0;
}
