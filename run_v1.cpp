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
int x, y, elevel, leader;
int group_members[9];

void count(int rank) {
    x = initialNodeConfig[rank][0];
    y = initialNodeConfig[rank][1];
    elevel = initialNodeConfig[rank][2];
    cout<< "Running thread: " << rank << endl;
    while(elevel > 0) {
        // cout << "Thread: " << rank << " energy level is: " << elevel<< endl;
        // Perform some computation here (e.g., use rank for different tasks)
        // Sleep 1 second
        this_thread::sleep_for(chrono::milliseconds(1000));
        elevel --;
    }
}

void leader_election(int rank, int node_details[4]){
    // If leader undefined
    // if(leader == -1){
        MPI_Bcast(node_details, 4, MPI_INT, rank, MPI_COMM_WORLD);
    //}
    
}

void main_flow(int rank) {
    int node_details[4];
    x = initialNodeConfig[rank][0];
    y = initialNodeConfig[rank][1];
    elevel = initialNodeConfig[rank][2];

    if (leader == -1 && rank == 0 ) {
        cout << "Leader not defined, found by: " << rank << endl;
        node_details[0] = x;
        node_details[1] = y;
        node_details[2] = elevel;
        node_details[3] = rank;
        leader_election(rank, node_details);
    }

    cout << "Rank " << rank << " received data=> " << "From: " << node_details[3] << " Energy level: " << node_details[2] << endl;
}

int main(int argc, char** argv) {
    leader = -1;

    // Initialize MPI environment
    MPI_Init_thread(&argc, &argv, MPI_THREAD_SERIALIZED, nullptr);

    int rank, size;
    MPI_Comm_rank(MPI_COMM_WORLD, &rank); // Get the process ID (rank)
    MPI_Comm_size(MPI_COMM_WORLD, &size); // Get the total number of processes

    // thread t1(count, rank);
    thread t2(main_flow, rank);

    // t1.join();
    t2.join();
    
    // Finalize MPI environment
    MPI_Finalize();
    return 0;
}