## Input
The following is an example tutorial on the expected input XML data:

### Example 1
The project has two tasks, "T1" and "T2," with durations of three and five days, respectively:

```xml
<project>
    <tasks>
        <task id="T1">
            <duration>3</duration>
        </task>
        <task id="T2">
            <duration>5</duration>
        </task>
</project>
```
### Example 2
The project has two tasks, "T1" and "T2," with durations of three and five days, respectively, and an FS dependency from T1 to T2 (T1 must finish before T2 can begin):

```xml
<project>
    <tasks>
        <task id="T1">
            <duration>3</duration>
            <dependencies>
                <dependency dependent-task-id="T2" type="FS"/>
                <!--dependency type may be FS, SS, FF or SF)-->
            </dependencies>
        </task>
        <task id="T2">
            <duration>5</duration>
        </task>
</project>
```
### Example 3
The project has two tasks, "T1" and "T2," with durations of three and five days, respectively, and an FS dependency from T1 to T2 (T1 must finish before T2 can begin). There are also three trucks available.  The tasks T1 and T2 require the usage of one and two trucks, respectively.

```xml
<project>
    <resources>
        <resource id="TRUCK" capacity="3"/>
    </resources>
    <tasks>
        <task id="T1">
            <duration>3</duration>
            <dependencies>
                <dependency dependent-task-id="T2" type="FS"/>
                <!--dependency type may be FS, SS, FF or SF)-->
            </dependencies>
            <allocations>
                <allocation resource-id="TRUCK" level="1"/>
            </allocations>
        </task>
        <task id="T2">
            <duration>5</duration>
            <allocations>
                <allocation resource-id="TRUCK" level="2"/>
            </allocations>
        </task>
</project>
```
### Example 4
The project has two tasks, "T1" and "T2," with durations of three and five days, respectively, and an FS dependency from T1 to T2 (T1 must finish before T2 can begin). There are also three trucks available.  The tasks T1 and T2 require the usage of one and two trucks, respectively. The task T1 requires the whole effort of worker John Doe, but the task T2 just requires 20% of his effort.
```xml
<project>
    <resources>
        <resource id="TRUCK" capacity="3"/>
        <resource id="JD" capacity="100"/>
    </resources>
    <tasks>
        <task id="T1">
            <duration>3</duration>
            <dependencies>
                <dependency dependent-task-id="T2" type="FS"/>
                <!--dependency type may be FS, SS, FF or SF)-->
            </dependencies>
            <allocations>
                <allocation resource-id="TRUCK" level="1"/>
                <allocation resource-id="JD" level="100"/>
            </allocations>
        </task>
        <task id="T2">
            <duration>5</duration>
            <allocations>
                <allocation resource-id="TRUCK" level="2"/>
                <allocation resource-id="JD" level="20"/>
            </allocations>
        </task>
</project>
```
### Example 5
The same as before, but considering the following calendar constraints:

- The project begins on July 1, 2024.
- Workdays are Monday through Saturday, with Sunday as a day off.
- Consider the holidays on January 1st and May 1st of 2024.
```xml
<project>
    <calendar>
        <kick-off-date>2024-07-01</kick-off-date>
        <idle-dates>
            <idle-date>2024-01-01</idle-date>
            <idle-date>2024-05-01</idle-date>
        </idle-dates>
        <idle-week-days>
            <idle-week-day>sunday</idle-week-day>
        </idle-week-days>
    </calendar>
    <resources>
        <resource id="TRUCK" capacity="3"/>
        <resource id="JD" capacity="100"/>
    </resources>
    <tasks>
        <task id="T1">
            <duration>3</duration>
            <dependencies>
                <dependency dependent-task-id="T2" type="FS"/>
                <!--dependency type may be FS, SS, FF or SF)-->
            </dependencies>
            <allocations>
                <allocation resource-id="TRUCK" level="1"/>
                <allocation resource-id="JD" level="100"/>
            </allocations>
        </task>
        <task id="T2">
            <duration>5</duration>
            <allocations>
                <allocation resource-id="TRUCK" level="2"/>
                <allocation resource-id="JD" level="20"/>
            </allocations>
        </task>
</project>
```
## Output
Upon normal termination (no input XML errors, for example) the return consists of XML data including the following tags:

|Tag|Scope|Description|
|--|--|--|
|makespan|Unique, global|The number of workdays required to complete the project|
|start-date|One per task|The task's start date, in standard ISO format (YYYY-MM-DD)|
|start-t|One per task|The number of workdays preceding the task's start date (zero if the task starts on the kick-off date)|
|finish-date|One per task|The task's finish date, in standard ISO format (YYYY-MM-DD)|
|finish-t|One per task|The number of workdays until the task is done (for the last tasks to finish, this value equals the value of the *makespan* tag minus one)|
