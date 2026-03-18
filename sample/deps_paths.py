import json
import argparse
import logging
import queue

logger = logging.getLogger(__name__)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    parser = argparse.ArgumentParser(description="Gets paths of the modules from deps.txt")
    parser.add_argument("json_file", help="Path to the module-info.json file")
    parser.add_argument("--deps-file", help="Path to the deps.txt file containing module keys (one per line)")

    args = parser.parse_args()

    with open(args.deps_file) as f:
        modules = f.read().splitlines()

    logger.info(f"Found {len(modules)} modules")

    with open(args.json_file) as f:
        module_info = json.load(f)

    logger.info("Loaded module-info.json")

    found_paths = set()

    for module in modules:
        module_data = module_info.get(module)
        if module_data is None:
            logger.warning(f"Module {module} not found in module-info.json")
            continue

        paths = set(module_data.get("path", []))
        if len(paths) == 0:
            logger.warning(f"Path for module {module} is not set or empty")
            continue

        new_paths = paths - found_paths
        logger.info(f"{module} found {len(new_paths)} new paths")

        found_paths.update(paths)
        for path in new_paths:
            print(path)

    logger.info("Done")
