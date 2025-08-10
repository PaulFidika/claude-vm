Right now, we will be a dev-container / conversational-orchestrator, but in the future we will want to move up the abstraction tree, to creating a task-orchestrator. The task-orchestration platform will use our container / conversational layer.

### Dev Lifecycle

*Plan. Develop. Test. Deploy.*

Ideally, we want to incorporate this entire loop. In a containeriazed environment, a single coding-agent could go through all of these steps (if it has access to the deployment pipeline).

### Employee Competency

The more compotent your employee is, the less managerial oversight is needed. For example, if you have a high-competency engineer, you can give them a general outline of what you want, and the constraints they need to meet, and they'll figure out the rest. For a junior engineer, you'll have to do a lot more upfront planning, and also spend more time reviewing his / her work.

AI-agents and humans both fall along this spectrum; the more incompetent your employees / agents, the more planning / QA you'll need to do.

Do not do more planning / QA than needed.

### Well Written Tasks

High-Competency Lifecycle:
- Give a rough specification of what you want done.
- Explain why it matters / what business goal you're trying to achieve with this task. This gives the agent context; the agent may decide you gave it the wrong task for the goal; it might find a much simpler / easier task that satisfies the same goal.
- Do not communicate further.
- Get notified when the task is done.

Low Competency Lifecycle:
- Clearly specify the desired endstate of the task.
- Link to all resources that would be helpful with the task.
- Give a detailed checklist of all steps necessary to complete the task.
- Give success / failure metrics that the agent can use to sanity-test their progress on their own. Examples:
    - if the PDF is less than 3 pages long, you probably messed up,
    - if the go binary does not build, you probably messed up
    - if the integration tests fail, you probably messed up,
    - if you cannot use puppeteer to login via the UI, you probably messed up...
- Check in with each other frequently (agent asks clarifying questions)
- Review all completed work; ask for revisions.

