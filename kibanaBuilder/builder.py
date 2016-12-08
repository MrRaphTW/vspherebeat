import os
import re
import argparse

parser = argparse.ArgumentParser(description='Will read one directory to build the kibana files in another directory')
parser.add_argument('--source', help='Source folder where to find the json folders and files.')
parser.add_argument('--dest', help='Destination folder where to send resolved json files.')
args = parser.parse_args()

def parseFolder(thePath):
    try:
        dirList = os.listdir(thePath)
    except FileNotFoundError:
        print('ERROR - We have tried to explore an invalid folder [%s]' % thePath)
        exit(1)
    for elt in dirList:
        newPath = thePath + elt
        if os.path.isdir(newPath) == True:
            newPath += '/'
            #print("dir is : %s" % newPath)
            if elt.split('.')[-1] == 'json':
                handleJsonFolder(newPath)
            else:
                parseFolder(newPath)
        #else:
            #print("file is : %s" % newPath)

def resolveVariables(myFile, thePath):
    newLines = []
    for line in myFile.readlines():
        includeNameList = re.findall("{{(.*)}}", line)
        for includeName in includeNameList:
            try:
                includeFile = open(thePath + includeName)
            except FileNotFoundError:
                print("The mentionned file does not exist [%s]. Maybe a typo" % thePath + includeName)
            toInclude = resolveVariables(includeFile, thePath).replace("\n"," ").replace('\\', '\\\\').replace('\"', '\\\"')
            includeFile.close()
            toInclude = re.sub(r'\s\s', '', toInclude)
            line = line.replace('{{'+includeName+'}}',toInclude)
        newLines += line
    return ''.join(newLines)

def getNewPathName(thePath):
    newPath = args.dest
    newPath += thePath.replace(args.source,'')
    newPath = newPath[:-1]
    return newPath

def getPath(thePathName):
    name = thePathName.split('/')[-1]
    path = thePathName.replace(name,'')
    return path

def handleJsonFolder(thePath):
    jsonName = thePath.split("/")[-2]
    myFile = open(thePath + jsonName)
    resolved = resolveVariables(myFile, thePath)
    myFile.close()
    #print(resolved)
    newPathName = getNewPathName(thePath)
    newPath = getPath(newPathName)
    os.makedirs(newPath, exist_ok=True)
    targetFile = open(newPathName,'w+')
    targetFile.writelines(resolved)
    targetFile.close()

if args.source[-1] != '/':
    args.source += '/'
if args.dest[-2] != '/':
    args.dest += '/'
parseFolder(args.source)
