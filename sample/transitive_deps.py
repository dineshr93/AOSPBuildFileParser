import json
import argparse
import logging
import queue

logger = logging.getLogger(__name__)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    parser = argparse.ArgumentParser(description="Gets all transitive dependencies of the modules from deps.txt")
    parser.add_argument("json_file", help="Path to the module-info.json file")
    parser.add_argument("--deps-file", help="Path to the deps.txt file containing module keys (one per line)")

    args = parser.parse_args()

    with open(args.deps_file) as f:
        modules = f.read().splitlines()

    logger.info(f"Found {len(modules)} initial modules")

    with open(args.json_file) as f:
        module_info = json.load(f)

    logger.info("Loaded module-info.json")

    visited_modules = set()
    to_visit = queue.Queue()

    for module in modules:
        to_visit.put(module, block=False)

    while not to_visit.empty():
        module = to_visit.get(block=False)
        if module in visited_modules:
            # NOTE: happens if it was found multiple times before visiting it
            continue
        visited_modules.add(module)
        print(module)

        module_data = module_info.get(module)
        if module_data is None:
            logger.warning(f"Module {module} not found in module-info.json")
            continue

        dependencies = set(module_data.get("dependencies", []))
        # dependencies.update(module_data.get("shared_libs", []))
        # dependencies.update(module_data.get("static_libs", []))

        new_dependencies = dependencies - visited_modules
        logger.info(f"{module} found {len(new_dependencies)} new dependencies")
        for module in new_dependencies:
            to_visit.put(module, block=False)

    logger.info(f"Visited total of {len(visited_modules)} modules")

    logger.info("Done")
