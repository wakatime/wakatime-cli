#!/bin/bash

# Greet user and request their name
echo "The activity generator"
read -p "What is your name? " name

# Create an array of activities
activity[0]="Football"
activity[1]="Table Tennis"
activity[2]="8 Ball Pool"
activity[3]="PS5"
activity[4]="Blackjack"

array_length=${#activity[@]} # Store the length of the array
index=$(($RANDOM % $array_length)) # Randomly select an index from 0 to array_length

# Invite the user to join you participate in an activity
echo "Hi" $name, "would you like to play" ${activity[$index]}"?"
read -p "Answer: " answer
