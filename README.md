# MediaHub Data Source for Grafana

The MediaHub Data Source plugin allows you to seamlessly integrate **MediaHub OSS** with your Grafana dashboards. This plugin goes beyond standard numerical metrics, enabling you to query media metadata, securely stream high-fidelity files (video, audio, images), and monitor system audit logs directly within Grafana.

---

## Features

* **Rich Media Proxying:** Securely stream images, audio, and video files directly into Grafana panels without exposing your MediaHub API to the public internet. Supports HTTP Range requests for smooth video scrubbing.
* **Dynamic Metadata Tables:** Query and flatten complex metadata (including custom dynamic fields) into structured Grafana DataFrames.
* **Dashboard Variables:** Fully supports Grafana template variables for dynamic database selection and filtering.
* **Smart Content Handling:** Automatically detects content types (`mime_type`) and applies intelligent defaults (e.g., disabling preview generation for generic files to save bandwidth).
* **Audit Trail Monitoring:** Built-in support for fetching and displaying the MediaHub system audit log (requires Admin privileges).

---

## Configuration

To connect the data source to your MediaHub instance, navigate to **Connections > Data sources > Add new data source** and select **MediaHub**.

You will need the following information:

1. **MediaHub URL:** The base URL of your MediaHub instance (e.g., `https://mediahub.internal.com`).
2. **Username:** A valid OAuth2 / API Username generated within MediaHub.
3. **Password:** The corresponding password for authentication.

Click **Save & test**. The plugin will perform an authentication handshake and verify the connection.

---

## Query Models

The Query Editor provides three primary modes of operation depending on the data you wish to visualize:

### 1. Get Metadata Table
Retrieves a paginated list of entries from a specific database. 
* **Target:** Specify a search query (e.g., `artist = "Demo"`) or leave blank for all entries.
* **Preview Links:** For supported media databases (Image, Video, Audio), you can toggle the generation of a direct proxy link to display thumbnails or media in table cells.

### 2. Get Preview / Entry
Fetches a specific media file by its Entry ID. This is typically used with dynamic dashboard variables to display the details of a single selected file.
* **Target Selection:** Choose "Get ID" to query a specific entry ID.
* **Base64 Content:** Optionally encode the media directly into the data frame (best for small thumbnails). Leave this disabled to use the high-performance URL proxy for large files like video and audio.

### 3. Get Audit Logs
Retrieves the system audit trail. The data is formatted with a localized timestamp and a stringified JSON details column for easy viewing in a standard Table panel. *(Note: The authenticated user must have the `IsAdmin` global role in MediaHub to use this model).*

---

## Displaying Media in Dashboards

To render dynamic media (Images, Video, Audio, or Download buttons for generic files) directly inside a dashboard, we highly recommend using the **Business Text Panel** (formerly Dynamic Text Panel by Volkov Labs).

Because the MediaHub plugin automatically exposes the `mime_type` in the DataFrame, you can use pure Handlebars templating without enabling Grafana's `disable_sanitize_html` security override.

**Setup Instructions:**
1. Add a **Business Text Panel** to your dashboard.
2. Set your query to **Get Preview** (or use a Metadata table with preview links generated).
3. Paste the following snippet into the **Content (HTML)** editor:

```html
<div style="display: flex; justify-content: center; align-items: center; height: 100%; width: 100%;">
  
  {{#if (contains mime_type "video/")}}
    <video src="{{entry}}" controls style="max-width: 100%; max-height: 100%; object-fit: contain; border-radius: 8px;"></video>
  
  {{else if (contains mime_type "audio/")}}
    <audio src="{{entry}}" controls style="width: 100%; max-width: 400px;"></audio>
  
  {{else if (contains mime_type "image/")}}
    <img src="{{entry}}" style="max-width: 100%; max-height: 100%; object-fit: contain; border-radius: 8px;" />
  
  {{else}}
    <a href="{{entry}}" target="_blank" style="padding: 10px 20px; background-color: #3274D9; color: white; border-radius: 4px; text-decoration: none; font-weight: bold; font-family: sans-serif;">
      Download File
    </a>
  {{/if}}

</div>
```