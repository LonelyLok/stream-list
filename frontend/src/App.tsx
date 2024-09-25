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
  CircularProgress,
  Tooltip,
  Button,
} from '@mui/joy';

const backendUrl = 'https://stream-list-backend.onrender.com/upcoming_streams';

type Video = {
  id: string;
  scheduledStartTime: string;
  thumbnails: Record<string, any>;
  title: string;
  liveBroadcastContent?: string;
  streamer?: string;
};

type Data = {
  channelId: string;
  name: string;
  iconURL: string;
  videos: Video[];
};

function App() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [streamers, setStreamers] = useState<Omit<Data, 'videos'>[]>([]);
  const [loading, setLoading] = useState(true);
  const chipColorMap = {
    live: 'danger' as const,
    none: 'neutral' as const,
    upcoming: 'warning' as const,
  };
  const getData = async () => {
    try {
      const response = await fetch(`${backendUrl}`);
      const data: Record<string, Data> = await response.json();
      // await new Promise((resolve) => setTimeout(resolve, 2000));
      // const data: Record<string, Data> = testData;
      const videos = Object.values(data).flatMap((d) => {
        return d.videos.map((v) => ({ ...v, streamer: d.name }));
      });
      setLoading(false);
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
        <Tooltip title={'View source code on GitHub'}>
          <Button
            variant='plain'
            component='a'
            href='https://github.com/LonelyLok/stream-list'
            target='_blank'
            rel='noopener noreferrer'
            sx={{
              ml: 1,
              '&:hover': {
                backgroundColor: 'rgba(255, 255, 255, 0.1)', // Light white for hover effect
              },
            }}
          >
            <img
              style={{ width: 24, height: 24 }}
              src={
                'https://github.com/LonelyLok/LonelyLok.github.io/blob/master/src/assets/github-mark-white.png?raw=true'
              }
              alt='github'
            />
          </Button>
        </Tooltip>
      </Typography>
      {loading ? (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <CircularProgress />
          <Typography level='body-sm' sx={{ mt: 1, color: 'white' }}>
            Can take a up to a minute to load
          </Typography>
        </Box>
      ) : (
        <div>
          <Box
            sx={{ width: '100%', maxWidth: '1500px', margin: '0 auto', mb: 3 }}
          >
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
                <Tooltip title={streamer.name}>
                  <Avatar
                    key={index}
                    src={streamer.iconURL}
                    alt={`Streamer ${index + 1}`}
                    sx={{ width: 80, height: 80 }}
                  />
                </Tooltip>
              ))}
            </Stack>
          </Box>
          <Box sx={{ width: '100%', maxWidth: '1500px', margin: '0 auto' }}>
            <Grid container spacing={3} sx={{ flexGrow: 1 }}>
              {videos.map((video, index) => (
                <Grid xs={12} sm={3} md={4} lg={3} key={index}>
                  <Card
                    variant='outlined'
                    sx={{ height: 270, width: 320, p: 1 }}
                  >
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
                      <Tooltip title={video.title}>
                        <Typography level='title-lg' noWrap={true}>
                          {video.title}
                        </Typography>
                      </Tooltip>
                      <Typography level='body-sm' noWrap={true}>
                        Streamer: {video?.streamer}
                      </Typography>
                      <Typography level='body-sm'>
                        Time:{' '}
                        {new Date(video.scheduledStartTime).toLocaleString()}
                      </Typography>
                      <Typography level='body-sm'>
                        Status:
                        <Chip
                          size='sm'
                          variant='solid'
                          color={
                            video?.liveBroadcastContent
                              ? chipColorMap[
                                  video.liveBroadcastContent as
                                    | 'live'
                                    | 'none'
                                    | 'upcoming'
                                ]
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
        </div>
      )}
    </Box>
  );
}

export default App;
