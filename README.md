
#Overview
GoProj is a streamlined cloud-based project scheduling engine. Its purpose is to integrate with project management tools in order to provide them with autoscheduling capabilities as simply as possible. Given project constraints such as task dependencies, resource capacities, and task demands, it establishes task start dates so that all constraints are met and the project is completed in the shortest possible timeframe.

REST web services and XML data exchange are used for the integration. Input must include task durations, task dependencies, resource capacities, and task resource consumption. The output includes the minimum feasible timeframe attained and the corresponding task dates, so that task dependencies are met and the total resource consumption by tasks does not exceed any resource capacity at any time.
