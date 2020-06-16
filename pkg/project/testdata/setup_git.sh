#!/bin/sh

################################################################################
# Script Name  : setup_git.sh 
# Description  : Create a git repository to the given path
# Args
# 1.           : Path to the main directory
# 2.           : Option to be called (basic, worktree, submodule, git_file)
################################################################################

path=$1
option=$2

project="wakatime-cli"

set_identity()
{
    git config user.name "wakatime-cli"
    git config user.email "wakatime-cli@wakatime.com"
}

initialize_git()
{
    # Create the main folder
    mkdir -p "$path/$project"

    # Change directory
    cd "$path/$project"
    
    # Initialize empty git repository
    git init

    # Create directory
    mkdir -p "$path/$project/src/pkg"

    # Create dummy file
    touch "$path/$project/src/pkg/project.go"

    # Commit to generate history
    git add .
    git commit -m "initial commit"
}

initialize_git_worktree()
{
    # # Create the main folder
    # mkdir -p "$path/$project"

    # # Change directory
    # cd "$path/$project"

    # ###########################
    # # Create by git init      #
    # ###########################
    # # Initialize empty git repository
    # git --git-dir repo --work-tree project init -q

    # # Create directory
    # mkdir -p "$path/$project/project/src/pkg"

    # # Create file
    # touch "$path/$project/project/src/pkg/project.go"

    # # Change directory to git repo
    # cd "$path/$project/repo"

    # # Commit to generate history
    # git add .
    # git commit -m "initial commit for worktree"

    ###########################
    # Create by git worktree  #
    ###########################
    # Change directory
    cd "$path/$project"

    # Add worktree
    git worktree add -b feature/api "../api"
}

initialize_git_submodule()
{
    # Create submodule directory
    mkdir -p "${path}/otherproject/module_a/src/lib"

    # Change directory
    cd "$path/otherproject/module_a"

    # Initialize empty git repository
    git init

    # Create file
    touch "src/lib/lib.cpp"

    # Commit to generate history
    git add .
    git commit -m "initial commit for submodule"

    # Change directory to main repo
    cd "$path/$project"

    # add submodule
    git submodule add "$path/otherproject/module_a" lib/module_a

    # Commit to generate history
    git add .
    git commit -m "add submodule module_a"
}

add_git_file()
{
    ###########################
    # Create by relative path #
    ###########################
    # Create other project directory
    mkdir -p "${path}/otherproject"

    # Change directory
    cd "$path/otherproject"

    # add .git file
    echo "gitdir: ../wakatime-cli/.git" >> ".git"

    # Create directory
    mkdir -p "$path/otherproject/src/pkg"
    
    # Create dummy file
    touch "$path/otherproject/src/pkg/project.go"

    ###########################
    # Create by absolute path #
    ###########################
    # Create some project directory
    mkdir -p "${path}/someproject"

    # Change directory
    cd "$path/someproject"

    # add .git file
    echo "gitdir: $path/wakatime-cli/.git" >> ".git"

    # Create directory
    mkdir -p "$path/someproject/src/pkg"
    
    # Create dummy file
    touch "$path/someproject/src/pkg/project.go"
}

# Always set identity
set_identity

case $option in
    "basic")
        initialize_git
        # Checkout to branch
        git checkout -b feature/detection
        ;;
    "worktree")
        initialize_git
        initialize_git_worktree
        cd "$path/$project"
        # Checkout to branch
        git checkout -b bugfix/log
        ;;
    "submodule")
        initialize_git
        initialize_git_submodule
        cd "$path/$project"
        # Checkout to branch
        git checkout -b bugfix/log
        ;;
    "git_file")
        initialize_git
        add_git_file
        cd "$path/otherproject"
        # Checkout to branch
        git checkout -b feature/detection
        ;;
    *)
        echo "invalid option"
        ;;
esac

exit 0
