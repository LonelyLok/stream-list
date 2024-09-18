import { useState, useEffect } from 'react';
// import { testData } from './test-data/get-data';
import {
  Box,
  Grid,
  Typography,
  Card,
  CardOverflow,
  CardContent,
  AspectRatio,
  Link,
  Chip,
  Stack,
  Avatar,
} from '@mui/joy';

const backendUrl = 'https://stream-list-backend.onrender.com/upcoming_streams';

type Video = {
  id: string;
  scheduledStartTime: string;
  thumbnails: Record<string, any>;
  title: string;
  liveBroadcastContent?: string;
};

type Data = {
  channelId: string;
  name: string;
  iconURL: string;
  videos: Video[];
};

function App() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [streamers, setStreamers] = useState<Omit<Data,'videos'>[]>([]);
  const getData = async () => {
    try {
      const response = await fetch(`${backendUrl}`);
      const data: Record<string, Data> = await response.json();
      // const data: Record<string, Data> = testData;
      const videos = Object.values(data).flatMap((d) => d.videos);
      setStreamers(
        Object.values(data).map((d) => {
          return {
            name: d.name,
            iconURL: d.iconURL,
            channelId: d.channelId,
          };
        })
      );
      setVideos(
        videos.sort(
          (a: any, b: any) =>
            new Date(b.scheduledStartTime).getTime() -
            new Date(a.scheduledStartTime).getTime()
        )
      );
    } catch (err) {
      console.log(err);
    }
  };
  useEffect(() => {
    getData();
  }, []);
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
      <Typography
        level='h2'
        sx={{ color: 'white', mb: 2, alignSelf: 'center' }}
      >
        Upcoming Streams
      </Typography>
      <Box sx={{ width: '100%', maxWidth: '1500px', margin: '0 auto', mb: 3 }}>
        <Stack
          direction='row'
          spacing={1}
          sx={{
            flexWrap: 'wrap',
            justifyContent: 'center',
            gap: 1,
            mb: 2,
          }}
        >
          {streamers.map((streamer, index) => (
            <Avatar
              key={index}
              src={streamer.iconURL}
              alt={`Streamer ${index + 1}`}
              sx={{ width: 80, height: 80 }}
            />
          ))}
        </Stack>
      </Box>
      <Box sx={{ width: '100%', maxWidth: '1500px', margin: '0 auto' }}>
        <Grid container spacing={3} sx={{ flexGrow: 1 }}>
          {videos.map((video, index) => (
            <Grid xs={12} sm={3} md={4} lg={3} key={index}>
              <Card variant='outlined' sx={{ height: 240, width: 320, p: 1 }}>
                <AspectRatio ratio='2'>
                  <CardOverflow>
                    <Link
                      href={`https://www.youtube.com/watch?v=${video.id}`}
                      overlay
                      underline='none'
                      target='_blank'
                      rel='noopener noreferrer'
                    >
                      <img
                        src={video.thumbnails.medium.url}
                        height={video.thumbnails.medium.height}
                        width={video.thumbnails.medium.width}
                      />
                    </Link>
                  </CardOverflow>
                </AspectRatio>
                <CardContent>
                  <Typography level='title-lg' noWrap={true}>
                    {video.title}
                  </Typography>
                  <Typography level='body-sm'>
                    Time: {new Date(video.scheduledStartTime).toLocaleString()}
                  </Typography>
                  <Typography level='body-sm'>
                    Status:
                    <Chip
                      size='sm'
                      variant='solid'
                      color={
                        video?.liveBroadcastContent === 'live'
                          ? 'danger'
                          : 'neutral'
                      }
                      sx={{ marginLeft: '8px' }}
                    >
                      {video?.liveBroadcastContent?.toUpperCase()}
                    </Chip>
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Box>
    </Box>
  );
}

export default App;
