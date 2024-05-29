import re
import shutil
import os

OBSIDIAN_PATH = "/Users/williamchrisp/Documents/Notes/Cortex"
TMP_PAGES_PATH = "./tmp_pages"
PAGES_PATH = "./src/pages"
TMP_IMAGES_PATH = "./tmp_images"
IMAGES_PATH = "./public/images"
IGNORED_FILES = [".DS_Store"]

def copy_folder(source, destination):
    try:
        # Delete folder if it exists
        if os.path.exists(destination):
            shutil.rmtree(destination)
        os.makedirs(destination)
        
        # Loop through all files and subdirectories in the source folder
        for root, dirs, files in os.walk(source):
            # Construct the corresponding destination path
            relative_path = os.path.relpath(root, source)
            dest_dir = os.path.join(destination, relative_path)
            
            # Ensure the destination subdirectory exists
            if not os.path.exists(dest_dir):
                os.makedirs(dest_dir)
            
            # Copy each file
            for file in files:
                if file in IGNORED_FILES:
                    print(f"Ignoring {file}")
                    continue
                src_file = os.path.join(root, file)
                dest_file = os.path.join(dest_dir, file)
                shutil.copy2(src_file, dest_file)
                print(f"Copied {src_file} to {dest_file}")
                
    except Exception as e:
        print("An error occurred:", e)

def update_header_links(text):
    # Regular expression to find the pattern [Text](#Link)
    pattern = re.compile(r'\[([^\]]+)\]\(#([^\)]+)\)')
    
    # Function to replace %20 with hyphens and convert to lowercase
    def replace_func(match):
        text = match.group(1)
        link = match.group(2).replace('%20', ' ').lower().replace(' ', '-')
        return f'[{text}](#{link})'
    
    # Substitute the matches using the replace function
    updated_text = pattern.sub(replace_func, text)
    return updated_text

def update_image_links(text):
    # Regular expression to find the pattern ![AltText](path/to/image)
    pattern = re.compile(r'!\[([^\]]+)\]\((cortex/[^\)]+)\)')
    
    # Function to replace 'cortex/' with '/'
    def replace_func(match):
        alt_text = match.group(1)
        link = match.group(2).replace('cortex/', '/')
        return f'![{alt_text}]({link})'
    
    # Substitute the matches using the replace function
    updated_text = pattern.sub(replace_func, text)
    return updated_text

def update_page_links(text):
    # Regular expression to find the pattern [Text](cortex/pages/path/to/file.md)
    pattern = re.compile(r'\[([^\]]+)\]\((cortex/pages/[^\)]+)\.md\)')
    
    # Function to capitalize the link text and adjust the path
    def replace_func(match):
        text = match.group(1)
        link = re.sub(r'cortex/pages', '', match.group(2)).replace('.md', '')
        return f'[{text}]({link})'
    
    # Substitute the matches using the replace function
    updated_text = pattern.sub(replace_func, text)
    return updated_text

def check_publish(text):
    match = re.match(r'---\n(.*?)\n---', text, re.DOTALL)
    if match:
        front_matter = match.group(1)
        # Find the publish field in the front matter
        publish_match = re.search(r'^publish:\s*"(true|false)"', front_matter, re.MULTILINE)
        if publish_match:
            publish_value = publish_match.group(1).strip().lower() == 'true'
            return publish_value
        return False

def process_files_recursively(folder_path):
    # Walk through all files in the folder and its subfolders
    for root, directories, files in os.walk(folder_path):
        for filename in files:
            file_path = os.path.join(root, filename)

            # Read input file
            with open(file_path, 'r') as file:
                content = file.read()

            if not check_publish(content):
                os.remove(file_path)
                print(file_path)
                print(f"Ignoring {filename} - Metadata value publish is not true.")
                image_folder = file_path.replace(TMP_PAGES_PATH, TMP_IMAGES_PATH).replace(".md", "")
                if os.path.exists(image_folder):
                    shutil.rmtree(image_folder)
                print(f"Ignoring {image_folder} - Metadata value publish for relevant page is not true.")
                continue

            # Process the file
            print("Processing file:", file_path)
            updated_header_links_content = update_header_links(content)
            updated_image_links_content = update_image_links(updated_header_links_content)
            updated_page_links_content = update_page_links(updated_image_links_content)

            with open(file_path, 'w') as file:
                file.write(updated_page_links_content)
       
def main():
    print("Copying pages from obsidian")
    copy_folder(OBSIDIAN_PATH + "/pages", TMP_PAGES_PATH)
    print("Copying images from obsidian")
    copy_folder(OBSIDIAN_PATH + "/images", TMP_IMAGES_PATH)
    print("Processing Files.")
    process_files_recursively(TMP_PAGES_PATH)
    copy_folder(TMP_PAGES_PATH, PAGES_PATH)
    copy_folder(TMP_IMAGES_PATH, IMAGES_PATH)
    print("Cleaning up tmp folder")
    shutil.rmtree(TMP_PAGES_PATH)
    shutil.rmtree(TMP_IMAGES_PATH)
    print("Sync complete.")

main()